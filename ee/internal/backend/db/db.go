package db

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Transaction interface {
	DB
	Commit() error

	// Does nothing if already committed.
	Rollback() error
}

func LoggedRollback(z *zap.Logger, tx Transaction) {
	if err := tx.Rollback(); err != nil {
		z.Error("rollback error", zap.Error(err))
	}
}

type DB interface {
	Connect(context.Context) error
	Setup(context.Context) error
	Teardown(context.Context) error

	Debug() DB

	// Begina a transaction.
	Begin(context.Context) (Transaction, error)

	Transaction(context.Context, func(tx DB) error) error

	// Returns sdkerrors.ErrAlreadyExists if either id or name is duplicate.
	CreateOrg(context.Context, sdktypes.Org) error

	// Returns sdkerrors.ErrNotFound if id is not found.
	GetOrgByID(context.Context, sdktypes.OrgID) (sdktypes.Org, error)

	// Returns sdkerrors.ErrNotFound if name is not found.
	GetOrgByName(context.Context, sdktypes.Symbol) (sdktypes.Org, error)

	// Returns sdkerrors.ErrNotFound if user or org are not found.
	// Returns sdkerrors.ErrAlreadyExists if user is already a member at the org.
	AddUserOrgMembership(context.Context, sdktypes.OrgID, sdktypes.UserID) error

	// Does not return an error if user already is not a member of that org.
	RemoveUserOrgMembership(context.Context, sdktypes.OrgID, sdktypes.UserID) error

	// Returns sdkerrors.ErrNotFound if org is not found.
	ListOrgMemberships(context.Context, sdktypes.OrgID) ([]sdktypes.User, error)

	IsMember(context.Context, sdktypes.OrgID, sdktypes.UserID) (bool, error)

	// Returns sdkerrors.ErrAlreadyExists if either id or name is duplicate.
	CreateUser(context.Context, sdktypes.User) error

	// Returns sdkerrors.ErrNotFound if not found.
	GetUserByID(context.Context, sdktypes.UserID) (sdktypes.User, error)

	// Returns sdkerrors.ErrNotFound if not found.
	GetUserByName(context.Context, sdktypes.Symbol) (sdktypes.User, error)

	GetUserByExternalID(ctx context.Context, eid string) (sdktypes.User, error)
	AddExternalIDToUser(ctx context.Context, uid sdktypes.UserID, externalID, idType, idEmail string) error

	ListUserOrgsMemberships(context.Context, sdktypes.UserID) ([]sdktypes.Org, error)
}
