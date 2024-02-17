package sdktypes

const ConnectionIDKind = "connection"

type ConnectionID = *id[connectionIDTraits]

var _ ID = (ConnectionID)(nil)

type connectionIDTraits struct{}

func (connectionIDTraits) Kind() string                   { return ConnectionIDKind }
func (connectionIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseConnectionID(raw string) (ConnectionID, error) {
	return parseTypedID[connectionIDTraits](raw)
}

func StrictParseConnectionID(raw string) (ConnectionID, error) {
	return strictParseTypedID[connectionIDTraits](raw)
}

func NewConnectionID() ConnectionID { return newID[connectionIDTraits]() }
