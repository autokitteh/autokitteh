package users

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
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
	// This is useful for single tenants setups where all users belong to the same org.
	DefaultOrgID string `json:"default_org_id"`
}

func (c *Config) GetDefaultOrgID() (sdktypes.OrgID, error) {
	return sdktypes.ParseOrgID(c.DefaultOrgID)
}

var Configs = configset.Set[Config]{
	Default: &Config{},
}

type users struct {
	db  db.DB
	l   *zap.Logger
	cfg *Config
}

func New(cfg *Config, db db.DB, l *zap.Logger) (sdkservices.Users, error) {
	if _, err := cfg.GetDefaultOrgID(); err != nil {
		return nil, fmt.Errorf("invalid default org ID: %w", err)
	}

	return &users{cfg: cfg, db: db, l: l}, nil
}

func (u *users) Create(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error) {
	if user.ID().IsValid() {
		return sdktypes.InvalidUserID, sdkerrors.NewInvalidArgumentError("user ID must be empty")
	}

	if err := authz.CheckContext(ctx, sdktypes.InvalidUserID, "create:create", authz.WithData("user", user)); err != nil {
		return sdktypes.InvalidUserID, err
	}

	var uid sdktypes.UserID

	err := u.db.Transaction(ctx, func(db db.DB) error {
		var oid sdktypes.OrgID

		if !user.DefaultOrgID().IsValid() {
			// If user has not default org id set, set the one from the config, if specified.
			if oid, _ = u.cfg.GetDefaultOrgID(); !oid.IsValid() {
				// ... otherwise create a new personal org for that user.
				org := sdktypes.NewOrg()

				if user.DisplayName() != "" {
					org = org.WithDisplayName(fmt.Sprintf("%s's Personal Org", user.DisplayName()))
				}

				var err error
				if oid, err = orgs.Create(ctx, db, org); err != nil {
					return fmt.Errorf("create personal org: %w", err)
				}
			}
		}

		var err error
		if uid, err = db.CreateUser(ctx, user.WithNewID().WithDefaultOrgID(oid)); err != nil {
			return err
		}

		if oid.IsValid() {
			if err := orgs.AddMember(ctx, db, oid, uid); err != nil {
				return fmt.Errorf("add as member to personal org: %w", err)
			}
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
	); err != nil {
		return sdktypes.InvalidUser, err
	}

	return u.db.GetUser(ctx, id, email)
}

func (u *users) Update(ctx context.Context, user sdktypes.User) error {
	if !user.ID().IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing user ID")
	}

	if err := authz.CheckContext(ctx, user.ID(), "update:update", authz.WithData("user", user)); err != nil {
		return err
	}

	return u.db.UpdateUser(ctx, user)
}
