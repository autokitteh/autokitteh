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

var OrgAdminRoleName = sdktypes.NewSymbol("admin")

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

// This is called by the user, and will also add the user to the org as its admin.
func (o *orgs) Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error) {
	var oid sdktypes.OrgID

	err := o.db.Transaction(ctx, func(tx db.DB) (err error) {
		if oid, err = Create(ctx, tx, org); err != nil {
			return
		}

		m := sdktypes.NewOrgMember(oid, authcontext.GetAuthnUserID(ctx)).
			WithStatus(sdktypes.OrgMemberStatusActive).
			WithRoles(OrgAdminRoleName)

		if err = AddMember(ctx, tx, m); err != nil {
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

func (o *orgs) BatchGetByIDs(ctx context.Context, ids []sdktypes.OrgID) ([]sdktypes.Org, error) {
	for _, id := range ids {
		if err := authz.CheckContext(ctx, id, "read:get", authz.WithConvertForbiddenToNotFound); err != nil {
			return nil, err
		}
	}

	return o.db.BatchGetOrgs(ctx, ids)
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

func (o *orgs) ListMembers(ctx context.Context, id sdktypes.OrgID) ([]sdktypes.OrgMember, error) {
	if err := authz.CheckContext(ctx, id, "read:list-members"); err != nil {
		return nil, err
	}

	return o.db.ListOrgMembers(ctx, id)
}

// This is meant to be called from a transaction.
// This function does not require authentication! It is used by the user service to add a user to an org.
func AddMember(ctx context.Context, db db.DB, m sdktypes.OrgMember) error {
	return db.AddOrgMember(ctx, m)
}

func (o *orgs) AddMember(ctx context.Context, m sdktypes.OrgMember) error {
	if uid := m.UserID(); !uid.IsValid() {
		m = m.WithUserID(authcontext.GetAuthnUserID(ctx))
	}

	if s := m.Status(); s == sdktypes.OrgMemberStatusUnspecified {
		m = m.WithStatus(sdktypes.OrgMemberStatusInvited)
	}

	if err := authz.CheckContext(
		ctx,
		m.OrgID(),
		"write:add-member",
		authz.WithData("org_member", m),
		authz.WithAssociationWithID("user", m.UserID()),
	); err != nil {
		return err
	}

	return AddMember(ctx, o.db, m)
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

func (o *orgs) GetMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (sdktypes.OrgMember, error) {
	if !uid.IsValid() {
		uid = authcontext.GetAuthnUserID(ctx)
	}

	m, err := o.db.GetOrgMember(ctx, oid, uid)
	if err != nil {
		return sdktypes.InvalidOrgMember, err
	}

	if err := authz.CheckContext(
		ctx,
		oid,
		"read:get-member",
		authz.WithAssociationWithID("user", uid),
		authz.WithData("member_status", m.Status().String()),
		authz.WithConvertForbiddenToNotFound,
	); err != nil {
		return sdktypes.InvalidOrgMember, err
	}

	return m, err
}

func (o *orgs) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.OrgMember, error) {
	if !uid.IsValid() {
		uid = authcontext.GetAuthnUserID(ctx)
	}

	if err := authz.CheckContext(ctx, uid, "read:get-orgs", authz.WithConvertForbiddenToNotFound); err != nil {
		return nil, err
	}

	return o.db.GetOrgsForUser(ctx, uid)
}

func (o *orgs) UpdateMember(ctx context.Context, m sdktypes.OrgMember, fm *sdktypes.FieldMask) error {
	if uid := m.UserID(); !uid.IsValid() {
		m = m.WithUserID(authcontext.GetAuthnUserID(ctx))
	}

	oid, uid := m.OrgID(), m.UserID()

	return o.db.Transaction(ctx, func(tx db.DB) error {
		curr, err := tx.GetOrgMember(ctx, oid, uid)
		if err != nil {
			return err
		}

		if err := authz.CheckContext(
			ctx,
			oid,
			"write:update-member",
			authz.WithAssociationWithID("user", uid),
			authz.WithData("new_roles", m.Roles()),
			authz.WithData("new_status", m.Status().String()),
			authz.WithData("current_status", curr.Status().String()),
			authz.WithData("current_roles", curr.Roles()),
			authz.WithFieldMask(fm),
		); err != nil {
			return err
		}

		return tx.UpdateOrgMember(ctx, m, fm)
	})
}
