package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Orgs interface {
	Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error)
	GetByID(ctx context.Context, id sdktypes.OrgID) (sdktypes.Org, error)
	Delete(ctx context.Context, id sdktypes.OrgID) error
	Update(ctx context.Context, org sdktypes.Org, fieldMask *sdktypes.FieldMask) error
	ListMembers(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.UserID, error)
	AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error
	RemoveMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error
	IsMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (bool, error)
	GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.Org, error)
}
