package sdktypes

import (
	"go.jetify.com/typeid"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type EventDestinationID struct{ id[typeid.AnyPrefix] }

var InvalidEventDestinationID EventDestinationID

type concreteEventDestinationID interface {
	ConnectionID | TriggerID
	ID
}

func NewEventDestinationID[T concreteEventDestinationID](in T) EventDestinationID {
	parsed := typeid.Must(ParseID[id[typeid.AnyPrefix]](in.String()))
	return EventDestinationID{parsed}
}

func ParseEventDestinationID(s string) (EventDestinationID, error) {
	if s == "" {
		return InvalidEventDestinationID, nil
	}

	parsed, err := ParseID[id[typeid.AnyPrefix]](s)
	if err != nil {
		return InvalidEventDestinationID, err
	}

	switch parsed.Kind() {
	case triggerIDKind, connectionIDKind:
		return EventDestinationID{parsed}, nil
	default:
		return InvalidEventDestinationID, sdkerrors.NewInvalidArgumentError("invalid executor id")
	}
}

func (e EventDestinationID) ToConnectionID() ConnectionID {
	id, _ := ParseConnectionID(e.String())
	return id
}

func (e EventDestinationID) ToTriggerID() TriggerID {
	id, _ := ParseTriggerID(e.String())
	return id
}

func (e EventDestinationID) IsConnectionID() bool { return e.Kind() == envIDKind }
func (e EventDestinationID) IsTriggerID() bool    { return e.Kind() == connectionIDKind }

func (e EventDestinationID) AsID() ID { return e }
