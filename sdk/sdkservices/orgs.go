package sdkservices

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type UserIDWithMemberStatus struct {
	UserID sdktypes.UserID          `json:"user_id"`
	Status sdktypes.OrgMemberStatus `json:"status"`
}

func (s UserIDWithMemberStatus) String() string { return fmt.Sprintf("%v: %v", s.UserID, s.Status) }

type OrgWithMemberStatus struct {
	Org    sdktypes.Org             `json:"org"`
	Status sdktypes.OrgMemberStatus `json:"status"`
}

func (s OrgWithMemberStatus) String() string { return fmt.Sprintf("%v, %v", s.Org, s.Status) }

type Orgs interface {
	Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error)
	GetByID(ctx context.Context, id sdktypes.OrgID) (sdktypes.Org, error)
	Delete(ctx context.Context, id sdktypes.OrgID) error
	Update(ctx context.Context, org sdktypes.Org, fieldMask *sdktypes.FieldMask) error
	ListMembers(ctx context.Context, oid sdktypes.OrgID) ([]*UserIDWithMemberStatus, error)
	AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID, status sdktypes.OrgMemberStatus) error
	UpdateMemberStatus(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID, status sdktypes.OrgMemberStatus) error
	RemoveMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error
	GetMemberStatus(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (sdktypes.OrgMemberStatus, error)
	GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]*OrgWithMemberStatus, error)
}
