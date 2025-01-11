package orgs

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/catnames"
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

// This is meant to be called from within a transaction (see users.go), and should
// not add the user to the org, as it is being done by the caller.
func Create(ctx context.Context, db db.DB, org sdktypes.Org) (sdktypes.OrgID, error) {
	if org.ID().IsValid() {
		return sdktypes.InvalidOrgID, sdkerrors.NewInvalidArgumentError("org ID must be empty")
	}

	org = org.WithNewID()

	if !org.Name().IsValid() {
		n := strings.ReplaceAll(catnames.Generate(), " ", "_") + "_Org"
		org = org.WithName(sdktypes.NewSymbol(n))
	}

	if err := authz.CheckContext(ctx, sdktypes.InvalidOrgID, "create:create", authz.WithData("org", org)); err != nil {
		return sdktypes.InvalidOrgID, err
	}

	return db.CreateOrg(ctx, org)
}

// This is called by the user, and will also add the user to the org.
func (o *orgs) Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error) {
	var oid sdktypes.OrgID

	err := o.db.Transaction(ctx, func(tx db.DB) (err error) {
		if oid, err = Create(ctx, tx, org); err != nil {
			return
		}

		if err = AddMember(ctx, tx, oid, authcontext.GetAuthnUserID(ctx), sdktypes.OrgMemberStatusActive); err != nil {
			return
		}

		return
	})

	return oid, err
}

func (o *orgs) GetByID(ctx context.Context, id sdktypes.OrgID) (sdktypes.Org, error) {
	if err := authz.CheckContext(ctx, id, "read:get", authz.WithConvertForbiddenToNotFound); err != nil {
		return sdktypes.InvalidOrg, err
	}

	return o.db.GetOrg(ctx, id, sdktypes.InvalidSymbol)
}

func (o *orgs) GetByName(ctx context.Context, n sdktypes.Symbol) (sdktypes.Org, error) {
	org, err := o.db.GetOrg(ctx, sdktypes.InvalidOrgID, n)
	if err != nil {
		return sdktypes.InvalidOrg, err
	}

	if err := authz.CheckContext(ctx, org.ID(), "read:get", authz.WithConvertForbiddenToNotFound); err != nil {
		return sdktypes.InvalidOrg, err
	}

	return org, nil
}

func (o *orgs) Delete(ctx context.Context, id sdktypes.OrgID) error {
	if err := authz.CheckContext(ctx, id, "delete:delete", authz.WithConvertForbiddenToNotFound); err != nil {
		return err
	}

	return o.db.DeleteOrg(ctx, id)
}

func (o *orgs) Update(ctx context.Context, org sdktypes.Org, fieldMask *sdktypes.FieldMask) error {
	if !org.ID().IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing org ID")
	}

	if err := org.ValidateUpdateFieldMask(fieldMask); err != nil {
		return err
	}

	if err := authz.CheckContext(ctx, org.ID(), "update:update", authz.WithData("org", org), authz.WithFieldMask(fieldMask)); err != nil {
		return err
	}

	return o.db.UpdateOrg(ctx, org, fieldMask)
}

func (o *orgs) ListMembers(ctx context.Context, id sdktypes.OrgID) ([]*sdkservices.UserIDWithMemberStatus, error) {
	if err := authz.CheckContext(ctx, id, "read:list-members"); err != nil {
		return nil, err
	}

	return o.db.ListOrgMembers(ctx, id)
}

// This is meant to be called from a transaction.
// This function does not require authentication! It is used by the user service to add a user to an org.
func AddMember(ctx context.Context, db db.DB, oid sdktypes.OrgID, uid sdktypes.UserID, s sdktypes.OrgMemberStatus) error {
	return db.AddOrgMember(ctx, oid, uid, s)
}

func (o *orgs) AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID, s sdktypes.OrgMemberStatus) error {
	if !uid.IsValid() {
		uid = authcontext.GetAuthnUserID(ctx)
	}

	if s == sdktypes.OrgMemberStatusUnspecified {
		s = sdktypes.OrgMemberStatusInvited
	}

	if err := authz.CheckContext(
		ctx,
		oid,
		"write:add-member",
		authz.WithAssociationWithID("user", uid),
		authz.WithData("status", s.String()),
	); err != nil {
		return err
	}

	return AddMember(ctx, o.db, oid, uid, s)
}

func (o *orgs) RemoveMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	if !uid.IsValid() {
		uid = authcontext.GetAuthnUserID(ctx)
	}

	if err := authz.CheckContext(ctx, oid, "delete:remove-member", authz.WithAssociationWithID("user", uid)); err != nil {
		return err
	}

	return o.db.RemoveOrgMember(ctx, oid, uid)
}

func (o *orgs) GetMemberStatus(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (sdktypes.OrgMemberStatus, error) {
	if !uid.IsValid() {
		uid = authcontext.GetAuthnUserID(ctx)
	}

	if err := authz.CheckContext(
		ctx,
		uid,
		"read:get-org-member-status",
		authz.WithAssociationWithID("org", oid),
		authz.WithConvertForbiddenToNotFound,
	); err != nil {
		return sdktypes.OrgMemberStatusUnspecified, err
	}

	return o.db.GetOrgMemberStatus(ctx, oid, uid)
}

func (o *orgs) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]*sdkservices.OrgWithMemberStatus, error) {
	if !uid.IsValid() {
		uid = authcontext.GetAuthnUserID(ctx)
	}

	if err := authz.CheckContext(ctx, uid, "read:get-orgs", authz.WithConvertForbiddenToNotFound); err != nil {
		return nil, err
	}

	return o.db.GetOrgsForUser(ctx, uid)
}

func (o *orgs) UpdateMemberStatus(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID, status sdktypes.OrgMemberStatus) error {
	if !uid.IsValid() {
		uid = authcontext.GetAuthnUserID(ctx)
	}

	return o.db.Transaction(ctx, func(tx db.DB) error {
		curr, err := tx.GetOrgMemberStatus(ctx, oid, uid)
		if err != nil {
			return err
		}

		if err := authz.CheckContext(
			ctx,
			oid,
			"write:update-member-status",
			authz.WithAssociationWithID("user", uid),
			authz.WithData("new_status", status.String()),
			authz.WithData("current_status", curr.String()),
		); err != nil {
			return err
		}

		return tx.UpdateOrgMemberStatus(ctx, oid, uid, status)
	})
}
