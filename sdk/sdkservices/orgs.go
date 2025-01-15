package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Orgs interface {
	Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error)
	GetByID(ctx context.Context, id sdktypes.OrgID) (sdktypes.Org, error)
	GetByName(ctx context.Context, name sdktypes.Symbol) (sdktypes.Org, error)
	// If any id not found, it will be ignored.
	BatchGetByIDs(ctx context.Context, ids []sdktypes.OrgID) ([]sdktypes.Org, error)
	Delete(ctx context.Context, id sdktypes.OrgID) error
	Update(ctx context.Context, org sdktypes.Org, fieldMask *sdktypes.FieldMask) error
	ListMembers(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.OrgMember, error)
	AddMember(ctx context.Context, m sdktypes.OrgMember) error
	UpdateMember(ctx context.Context, m sdktypes.OrgMember, fm *sdktypes.FieldMask) error
	RemoveMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error
	GetMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (sdktypes.OrgMember, error)
	GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.OrgMember, error)
}
