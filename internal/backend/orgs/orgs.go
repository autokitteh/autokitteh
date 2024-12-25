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

func Create(ctx context.Context, db db.DB, org sdktypes.Org) (sdktypes.OrgID, error) {
	if org.ID().IsValid() {
		return sdktypes.InvalidOrgID, sdkerrors.NewInvalidArgumentError("org ID must be empty")
	}

	org = org.WithNewID()

	if err := authz.CheckContext(ctx, sdktypes.InvalidOrgID, "create:create", authz.WithData("org", org)); err != nil {
		return sdktypes.InvalidOrgID, err
	}

	return db.CreateOrg(ctx, org)
}

func (o *orgs) Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error) {
	return Create(ctx, o.db, org)
}

func (o *orgs) GetByID(ctx context.Context, id sdktypes.OrgID) (sdktypes.Org, error) {
	if err := authz.CheckContext(ctx, id, "read:get", authz.WithConvertForbiddenToNotFound); err != nil {
		return sdktypes.InvalidOrg, err
	}

	return o.db.GetOrg(ctx, id)
}

func (o *orgs) Update(ctx context.Context, org sdktypes.Org) error {
	if !org.ID().IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing org ID")
	}

	if err := authz.CheckContext(ctx, org.ID(), "update:update", authz.WithData("org", org)); err != nil {
		return err
	}

	return o.db.UpdateOrg(ctx, org)
}

func (o *orgs) ListMembers(ctx context.Context, id sdktypes.OrgID) ([]sdktypes.UserID, error) {
	if err := authz.CheckContext(ctx, id, "read:list-users"); err != nil {
		return nil, err
	}

	return o.db.ListOrgMembers(ctx, id)
}

func AddMember(ctx context.Context, db db.DB, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	if err := authz.CheckContext(ctx, oid, "write:add-user"); err != nil {
		return err
	}

	return db.AddOrgMember(ctx, oid, uid)
}

func (o *orgs) AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	return AddMember(ctx, o.db, oid, uid)
}

func (o *orgs) RemoveMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	if err := authz.CheckContext(ctx, oid, "write:rm-user"); err != nil {
		return err
	}

	return o.db.RemoveOrgMember(ctx, oid, uid)
}

func (o *orgs) IsMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (bool, error) {
	if err := authz.CheckContext(ctx, oid, "read:is-member"); err != nil {
		return false, err
	}

	if err := authz.CheckContext(ctx, uid, "read:is-org-member"); err != nil {
		return false, err
	}

	return o.db.IsOrgMember(ctx, oid, uid)
}

func (o *orgs) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.Org, error) {
	if err := authz.CheckContext(ctx, uid, "read:get-orgs-for-user"); err != nil {
		return nil, err
	}

	return o.db.GetOrgsForUser(ctx, uid)
}
