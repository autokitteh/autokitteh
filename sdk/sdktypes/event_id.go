package sdktypes

const EventIDKind = "event"

type EventID = *id[eventIDTraits]

var _ ID = (EventID)(nil)

type eventIDTraits struct{}

func (eventIDTraits) Kind() string                   { return EventIDKind }
func (eventIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseEventID(raw string) (EventID, error) { return parseTypedID[eventIDTraits](raw) }

func StrictParseEventID(raw string) (EventID, error) { return strictParseTypedID[eventIDTraits](raw) }

func NewEventID() EventID { return newID[eventIDTraits]() }
