package sdktypes

import (
	"go.jetify.com/typeid"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type OwnerID struct{ id[typeid.AnyPrefix] }

var InvalidOwnerID OwnerID

type concreteOwnerID interface {
	UserID | OrgID
	ID
}

func NewOwnerID[T concreteOwnerID](in T) OwnerID {
	parsed := typeid.Must(ParseID[id[typeid.AnyPrefix]](in.String()))
	return OwnerID{parsed}
}

func ParseOwnerID(s string) (OwnerID, error) {
	if s == "" {
		return InvalidOwnerID, nil
	}

	parsed, err := ParseID[id[typeid.AnyPrefix]](s)
	if err != nil {
		return InvalidOwnerID, err
	}

	switch parsed.Kind() {
	case UserIDKind:
		return OwnerID{parsed}, nil
	default:
		return InvalidOwnerID, sdkerrors.NewInvalidArgumentError("invalid owner id")
	}
}

func (e OwnerID) ToUserID() UserID { id, _ := ParseUserID(e.String()); return id }
func (e OwnerID) ToOrgID() OrgID   { id, _ := ParseOrgID(e.String()); return id }

func (e OwnerID) IsUserID() bool { return e.Kind() == UserIDKind }
func (e OwnerID) IsOrgID() bool  { return e.Kind() == OrgIDKind }

func (e OwnerID) AsID() ID { return e }
