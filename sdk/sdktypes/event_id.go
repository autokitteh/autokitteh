package sdktypes

const EventIDKind = "evt"

type EventID = id[eventIDTraits]

type eventIDTraits struct{}

func (eventIDTraits) Prefix() string { return EventIDKind }

func NewEventID() EventID                    { return newID[EventID]() }
func ParseEventID(s string) (EventID, error) { return ParseID[EventID](s) }

var InvalidEventID EventID
