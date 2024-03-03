package sdktypes

const connectionIDKind = "connection"

type ConnectionID = id[connectionIDTraits]

type connectionIDTraits struct{}

func (connectionIDTraits) Prefix() string { return connectionIDKind }

func NewConnectionID() ConnectionID                          { return newID[ConnectionID]() }
func ParseConnectionID(s string) (ConnectionID, error)       { return ParseID[ConnectionID](s) }
func StrictParseConnectionID(s string) (ConnectionID, error) { return Strict(ParseConnectionID(s)) }

func IsConnectionID(s string) bool { return IsIDOf[connectionIDTraits](s) }

var InvalidConnectionID ConnectionID
