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
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	// If set, do not create a personal org for new users. Instead, set this as their default org.
	DefaultOrgID sdktypes.OrgID `json:"default_org_id"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
}

type users struct {
	db  db.DB
	l   *zap.Logger
	cfg *Config
}

func New(cfg *Config, db db.DB, l *zap.Logger) sdkservices.Users {
	return &users{cfg: cfg, db: db, l: l}
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
		// TODO: user or org name might already be taken, figure out a new org name.

		var oid sdktypes.OrgID

		if !user.DefaultOrgID().IsValid() {
			if u.cfg.DefaultOrgID.IsValid() {
				oid = u.cfg.DefaultOrgID
			} else {
				orgName := kittehs.Must1(sdktypes.ParseSymbol(fmt.Sprintf("%s_org", user.Name())))
				org := sdktypes.NewOrg(orgName).WithDisplayName(fmt.Sprintf("%s's Personal Org", user.DisplayName()))

				var err error
				if oid, err = orgs.Create(ctx, db, org); err != nil {
					return fmt.Errorf("create personal org: %w", err)
				}
			}
		}

		uid, err := db.CreateUser(ctx, user.WithNewID().WithDefaultOrgID(oid))
		if err != nil {
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

func (u *users) Get(ctx context.Context, id sdktypes.UserID, name sdktypes.Symbol, email string) (sdktypes.User, error) {
	if email == "" && !id.IsValid() {
		id = authcontext.GetAuthnUserID(ctx)
	}

	if err := authz.CheckContext(
		ctx,
		id,
		"read:get",
		authz.WithData("id", id.String()),
		authz.WithData("email", email),
		authz.WithData("name", name.String()),
	); err != nil {
		return sdktypes.InvalidUser, err
	}

	return u.db.GetUser(ctx, id, name, email)
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
