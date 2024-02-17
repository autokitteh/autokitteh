package sdktypes

const SessionIDKind = "s"

type SessionID = *id[SessionIDTraits]

var _ ID = (SessionID)(nil)

type SessionIDTraits struct{}

func (SessionIDTraits) Kind() string                   { return SessionIDKind }
func (SessionIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseSessionID(raw string) (SessionID, error) {
	return parseTypedID[SessionIDTraits](raw)
}

func StrictParseSessionID(raw string) (SessionID, error) {
	return strictParseTypedID[SessionIDTraits](raw)
}

func NewSessionID() SessionID { return newID[SessionIDTraits]() }
