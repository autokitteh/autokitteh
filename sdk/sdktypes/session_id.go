package sdktypes

const SessionIDKind = "ses"

type SessionID = id[sessionIDTraits]

type sessionIDTraits struct{}

func (sessionIDTraits) Prefix() string { return SessionIDKind }

func NewSessionID() SessionID                          { return newID[SessionID]() }
func ParseSessionID(s string) (SessionID, error)       { return ParseID[SessionID](s) }
func StrictParseSessionID(s string) (SessionID, error) { return Strict(ParseSessionID(s)) }

var InvalidSessionID SessionID
