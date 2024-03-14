package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Orgs interface {
	Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error)
	GetByID(ctx context.Context, orgID sdktypes.OrgID) (sdktypes.Org, error)
	GetByName(ctx context.Context, name sdktypes.Symbol) (sdktypes.Org, error)
	AddMember(ctx context.Context, orgID sdktypes.OrgID, userID sdktypes.UserID) error
	RemoveMember(ctx context.Context, orgID sdktypes.OrgID, userID sdktypes.UserID) error
	ListMembers(ctx context.Context, orgID sdktypes.OrgID) ([]sdktypes.User, error) // TODO: pagination.
	IsMember(ctx context.Context, orgID sdktypes.OrgID, userID sdktypes.UserID) (bool, error)
	ListUserMemberships(ctx context.Context, userID sdktypes.UserID) ([]sdktypes.Org, error)
}
