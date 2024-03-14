package orgs

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/db"
)

type orgs struct {
	z  *zap.Logger
	db db.DB
}

func New(z *zap.Logger, db db.DB) sdkservices.Orgs {
	return wrap(&orgs{db: db, z: z}, db)
}

func (o *orgs) Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error) {
	org = org.WithID(sdktypes.NewOrgID())

	if err := org.Strict(); err != nil {
		return sdktypes.InvalidOrgID, err
	}

	tx, err := o.db.Begin(ctx)
	if err != nil {
		return sdktypes.InvalidOrgID, fmt.Errorf("db.begin: %w", err)
	}

	defer db.LoggedRollback(o.z, tx)

	if _, err := tx.GetUserByName(ctx, org.Name()); err == nil {
		return sdktypes.InvalidOrgID, fmt.Errorf("%w: a user with the same name already exists", sdkerrors.ErrAlreadyExists)
	} else if !errors.Is(err, sdkerrors.ErrNotFound) {
		return sdktypes.InvalidOrgID, fmt.Errorf("db.get_user_by_name: %w", err)
	}

	if err := tx.CreateOrg(ctx, org); err != nil {
		return sdktypes.InvalidOrgID, err
	}

	if err := tx.Commit(); err != nil {
		return sdktypes.InvalidOrgID, fmt.Errorf("db.commit: %w", err)
	}

	return org.ID(), nil
}

func (o *orgs) GetByID(ctx context.Context, oid sdktypes.OrgID) (sdktypes.Org, error) {
	return sdkerrors.IgnoreNotFoundErr(o.db.GetOrgByID(ctx, oid))
}

func (o *orgs) GetByName(ctx context.Context, h sdktypes.Symbol) (sdktypes.Org, error) {
	return sdkerrors.IgnoreNotFoundErr(o.db.GetOrgByName(ctx, h))
}

func (o *orgs) AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("db.begin: %w", err)
	}

	defer db.LoggedRollback(o.z, tx)

	if _, err := tx.GetOrgByID(ctx, oid); err != nil {
		if !errors.Is(err, sdkerrors.ErrNotFound) {
			return fmt.Errorf("%w: org not found", sdkerrors.ErrNotFound)
		}

		return fmt.Errorf("db.get_user_by_id: %w", err)
	}

	if _, err := tx.GetUserByID(ctx, uid); err != nil {
		if !errors.Is(err, sdkerrors.ErrNotFound) {
			return fmt.Errorf("%w: user not found", sdkerrors.ErrNotFound)
		}

		return fmt.Errorf("db.get_user_by_id: %w", err)

	}

	if err := tx.AddUserOrgMembership(ctx, oid, uid); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db.commit: %w", err)
	}

	return nil
}

func (o *orgs) RemoveMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	return o.db.RemoveUserOrgMembership(ctx, oid, uid)
}

func (o *orgs) ListMembers(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.User, error) {
	return o.db.ListOrgMemberships(ctx, oid)
}

func (o *orgs) IsMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (bool, error) {
	return o.db.IsMember(ctx, oid, uid)
}

func (o *orgs) ListUserMemberships(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.Org, error) {
	if !uid.IsValid() {
		return nil, sdkerrors.NewInvalidArgumentError("uid", "invalid user id")
	}

	return o.db.ListUserOrgsMemberships(ctx, uid)
}
