package orgs

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type orgs struct {
	db db.DB
	l  *zap.Logger
}

func New(db db.DB, l *zap.Logger) sdkservices.Orgs {
	return &orgs{db: db, l: l}
}

func (u *orgs) Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error) {
	if org.ID().IsValid() {
		return sdktypes.InvalidOrgID, sdkerrors.NewInvalidArgumentError("org ID must be empty")
	}

	if err := authz.CheckContext(ctx, sdktypes.InvalidOrgID, "create:create", authz.WithData("org", org)); err != nil {
		return sdktypes.InvalidOrgID, err
	}

	return u.db.CreateOrg(ctx, org)
}

func (u *orgs) GetByID(ctx context.Context, id sdktypes.OrgID) (sdktypes.Org, error) {
	if err := authz.CheckContext(ctx, id, "read:get"); err != nil {
		return sdktypes.InvalidOrg, err
	}

	return u.db.GetOrg(ctx, id, sdktypes.InvalidSymbol)
}

func (u *orgs) GetByName(ctx context.Context, name sdktypes.Symbol) (sdktypes.Org, error) {
	if err := authz.CheckContext(ctx, sdktypes.InvalidOrgID, "read:resolve"); err != nil {
		return sdktypes.InvalidOrg, err
	}

	o, err := u.db.GetOrg(ctx, sdktypes.InvalidOrgID, name)
	if err != nil {
		return sdktypes.InvalidOrg, err
	}

	if err := authz.CheckContext(ctx, o.ID(), "read:get"); err != nil {
		return sdktypes.InvalidOrg, err
	}

	return o, nil
}

func (u *orgs) Update(ctx context.Context, org sdktypes.Org) error {
	if !org.ID().IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing org ID")
	}

	if err := authz.CheckContext(ctx, org.ID(), "update:update", authz.WithData("org", org)); err != nil {
		return err
	}

	return u.db.UpdateOrg(ctx, org)
}

func (u *orgs) ListMembers(ctx context.Context, id sdktypes.OrgID) ([]sdktypes.UserID, error) {
	if err := authz.CheckContext(ctx, id, "read:list-users"); err != nil {
		return nil, err
	}

	return u.db.ListOrgMembers(ctx, id)
}

func (u *orgs) AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	if err := authz.CheckContext(ctx, oid, "write:add-user"); err != nil {
		return err
	}

	return u.db.AddOrgMember(ctx, oid, uid)
}

func (u *orgs) RemoveMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	if err := authz.CheckContext(ctx, oid, "write:rm-user"); err != nil {
		return err
	}

	return u.db.RemoveOrgMember(ctx, oid, uid)
}

func (u *orgs) IsMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (bool, error) {
	if err := authz.CheckContext(ctx, oid, "read:is-member"); err != nil {
		return false, err
	}

	if err := authz.CheckContext(ctx, uid, "read:is-org-member"); err != nil {
		return false, err
	}

	return u.db.IsOrgMember(ctx, oid, uid)
}

func (u *orgs) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.OrgID, error) {
	if err := authz.CheckContext(ctx, uid, "read:get-orgs-for-user"); err != nil {
		return nil, err
	}

	return u.db.GetOrgsForUser(ctx, uid)
}
