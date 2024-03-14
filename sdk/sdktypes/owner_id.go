package sdktypes

import (
	"go.jetpack.io/typeid"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type OwnerID struct{ id[typeid.AnyPrefix] }

type concreteOwnerID interface {
	UserID | OrgID
	ID
}

func NewOwnerID[T concreteOwnerID](in T) OwnerID {
	parsed := kittehs.Must1(ParseID[id[typeid.AnyPrefix]](in.String()))
	return OwnerID{parsed}
}

func ParseOwnerID(s string) (OwnerID, error) {
	parsed, err := ParseID[id[typeid.AnyPrefix]](s)
	if err != nil {
		return OwnerID{}, err
	}

	switch parsed.Kind() {
	case userIDKind, orgIDKind:
		return OwnerID{parsed}, nil
	default:
		return OwnerID{}, sdkerrors.NewInvalidArgumentError("invalid owner id")
	}
}

func (e OwnerID) ToUserID() UserID {
	id, _ := ParseUserID(e.String())
	return id
}

func (e OwnerID) ToOrgID() OrgID {
	id, _ := ParseOrgID(e.String())
	return id
}

func (e OwnerID) IsUserID() bool { return e.Kind() == userIDKind }
func (e OwnerID) IsOrgID() bool  { return e.Kind() == orgIDKind }
