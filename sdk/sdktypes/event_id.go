package sdktypes

const eventIDKind = "evt"

type EventID = id[eventIDTraits]

type eventIDTraits struct{}

func (eventIDTraits) Prefix() string { return eventIDKind }

func NewEventID() EventID                          { return newID[EventID]() }
func ParseEventID(s string) (EventID, error)       { return ParseID[EventID](s) }
func StrictParseEventID(s string) (EventID, error) { return Strict(ParseEventID(s)) }

var InvalidEventID EventID
