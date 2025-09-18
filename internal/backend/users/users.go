package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/orgs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	// If set, do not create a personal org for new users. Instead, set this as their default org.
	// This is useful for single-tenant setups where all users belong to the same org.
	DefaultOrgID string `koanf:"default_org_id"`

	UseDefaultUser bool `koanf:"use_default_user"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev:     &Config{UseDefaultUser: true},
}

type Users interface {
	sdkservices.Users

	Setup(context.Context) error

	HasDefaultUser() bool
}

type users struct {
	db  db.DB
	l   *zap.Logger
	cfg *Config

	defaultOrgID sdktypes.OrgID
}

func New(cfg *Config, db db.DB, l *zap.Logger) (Users, error) {
	oid, err := sdktypes.ParseOrgID(cfg.DefaultOrgID)
	if err != nil {
		return nil, fmt.Errorf("invalid default org ID: %w", err)
	}

	return &users{cfg: cfg, db: db, l: l, defaultOrgID: oid}, nil
}

func (u *users) HasDefaultUser() bool { return u.cfg.UseDefaultUser }

func (u *users) Setup(ctx context.Context) error {
	if u.cfg.UseDefaultUser {
		u.l.Warn("using default user")

		if _, err := u.db.GetUser(ctx, authusers.DefaultUser.ID(), ""); err == nil {
			// user exists - nothing to do.
			u.l.Info("default user exist in db")
			return nil
		} else if !errors.Is(err, sdkerrors.ErrNotFound) {
			return fmt.Errorf("get default user: %w", err)
		}

		// no user - populate.
		u.l.Info("no default user found in db, populating")

		var seed []sdktypes.Object

		if u.cfg.DefaultOrgID == "" {
			// Add both the default org and default user as the org admin.

			seed = []sdktypes.Object{
				authusers.DefaultUser,
				authusers.DefaultOrg,
				sdktypes.NewOrgMember(authusers.DefaultUser.DefaultOrgID(), authusers.DefaultUser.ID()).
					WithRoles(orgs.OrgAdminRoleName).
					WithStatus(sdktypes.OrgMemberStatusActive),
			}
		} else {
			// Add default user to the desired default org - do not make them an admin.

			seed = []sdktypes.Object{
				authusers.DefaultUser.WithDefaultOrgID(u.defaultOrgID),
				sdktypes.NewOrgMember(u.defaultOrgID, authusers.DefaultUser.ID()).
					WithStatus(sdktypes.OrgMemberStatusActive),
			}
		}

		if err := db.Populate(ctx, u.db, seed...); err != nil {
			return err
		}
	}

	return nil
}

func (u *users) Create(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error) {
	if user.ID().IsValid() {
		return sdktypes.InvalidUserID, sdkerrors.NewInvalidArgumentError("user ID must be empty")
	}

	if user.Status() == sdktypes.UserStatusUnspecified {
		user = user.WithStatus(sdktypes.UserStatusActive)
	}

	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidUserID,
		"create:create",
		authz.WithData("user", user),
		authz.WithData("status", user.Status().String()),
	); err != nil {
		return sdktypes.InvalidUserID, err
	}

	var uid sdktypes.UserID

	err := u.db.Transaction(ctx, func(db db.DB) error {
		var (
			err   error
			roles []sdktypes.Symbol
		)

		if oid := user.DefaultOrgID(); !oid.IsValid() {
			// If user has no default org id set, set the one from the config, if specified.
			if oid = u.defaultOrgID; !oid.IsValid() {
				// ... otherwise create a new personal org for that user.
				org := sdktypes.NewOrg()

				orgNamePrefix := user.DisplayName()
				if orgNamePrefix == "" {
					orgNamePrefix = strings.SplitN(user.Email(), "@", 2)[0]
				}

				org = org.WithDisplayName(orgNamePrefix + "'s Personal Org")

				var err error
				if oid, err = orgs.Create(ctx, db, org); err != nil {
					return fmt.Errorf("create personal org: %w", err)
				}

				// This is a new personal org, so the user is an admin.
				roles = []sdktypes.Symbol{orgs.OrgAdminRoleName}
			}

			// We will always have oid set by this point.
			user = user.WithDefaultOrgID(oid)
		}

		// We will always have user.DefaultOrgID set by this point.

		if uid, err = db.CreateUser(ctx, user); err != nil {
			return err
		}

		m := sdktypes.NewOrgMember(user.DefaultOrgID(), uid).WithStatus(sdktypes.OrgMemberStatusActive).WithRoles(roles...)

		if err := orgs.AddMember(ctx, db, m); err != nil {
			return fmt.Errorf("add as member to personal org: %w", err)
		}

		return nil
	})

	return uid, err
}

func (u *users) Get(ctx context.Context, id sdktypes.UserID, email string) (sdktypes.User, error) {
	if email == "" && !id.IsValid() {
		id = authcontext.GetAuthnUserID(ctx)
	}

	if err := authz.CheckContext(
		ctx,
		id,
		"read:get",
		authz.WithData("user_id", id.String()),
		authz.WithData("email", email),
		authz.WithConvertForbiddenToNotFound,
	); err != nil {
		return sdktypes.InvalidUser, err
	}

	return u.db.GetUser(ctx, id, email)
}

func (u *users) GetID(ctx context.Context, email string) (sdktypes.UserID, error) {
	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidUserID,
		"read:get-id",
		authz.WithData("email", email),
		authz.WithConvertForbiddenToNotFound,
	); err != nil {
		return sdktypes.InvalidUserID, err
	}

	r, err := u.db.GetUser(ctx, sdktypes.InvalidUserID, email)
	if err != nil {
		return sdktypes.InvalidUserID, err
	}

	return r.ID(), nil
}

func (u *users) Update(ctx context.Context, user sdktypes.User, fieldMask *sdktypes.FieldMask) error {
	if !user.ID().IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing user ID")
	}

	if err := user.ValidateUpdateFieldMask(fieldMask); err != nil {
		return err
	}

	if err := authz.CheckContext(
		ctx,
		user.ID(),
		"update:update",
		authz.WithData("user", user),
		authz.WithFieldMask(fieldMask),
		authz.WithData("status", user.Status().String()),
	); err != nil {
		return err
	}

	return u.db.UpdateUser(ctx, user, fieldMask)
}
