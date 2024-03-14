package orgs

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/db"
)

type auth struct {
	orgs sdkservices.Orgs
	db   db.DB
}

func wrap(in sdkservices.Orgs, db db.DB) sdkservices.Orgs { return &auth{orgs: in, db: db} }

func (o *auth) Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error) {
	userID := authcontext.GetAuthnUserID(ctx)

	orgID, err := o.orgs.Create(ctx, org)
	if err != nil {
		return sdktypes.InvalidOrgID, err
	}

	if userID.IsValid() {
		if err := o.db.AddUserOrgMembership(ctx, orgID, userID); err != nil {
			return sdktypes.InvalidOrgID, fmt.Errorf("add creator as member: %w", err)
		}
	}

	return orgID, nil
}

func (o *auth) GetByID(ctx context.Context, oid sdktypes.OrgID) (sdktypes.Org, error) {
	return o.orgs.GetByID(ctx, oid)
}

func (o *auth) GetByName(ctx context.Context, h sdktypes.Symbol) (sdktypes.Org, error) {
	return o.orgs.GetByName(ctx, h)
}

func (o *auth) AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	return o.orgs.AddMember(ctx, oid, uid)
}

func (o *auth) RemoveMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	return o.orgs.RemoveMember(ctx, oid, uid)
}

func (o *auth) ListMembers(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.User, error) {
	return o.orgs.ListMembers(ctx, oid)
}

func (o *auth) IsMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (bool, error) {
	return o.orgs.IsMember(ctx, oid, uid)
}

func (o *auth) ListUserMemberships(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.Org, error) {
	if !uid.IsValid() {
		uid = authcontext.GetAuthnUserID(ctx)
	}

	return o.orgs.ListUserMemberships(ctx, uid)
}
