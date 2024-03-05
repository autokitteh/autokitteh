package sdktypes

const envIDKind = "env"

type EnvID = id[envIDTraits]

type envIDTraits struct{}

func (envIDTraits) Prefix() string { return envIDKind }

func NewEnvID() EnvID                          { return newID[EnvID]() }
func ParseEnvID(s string) (EnvID, error)       { return ParseID[EnvID](s) }
func StrictParseEnvID(s string) (EnvID, error) { return Strict(ParseEnvID(s)) }

func IsEnvID(s string) bool { return IsIDOf[envIDTraits](s) }

var InvalidEnvID EnvID
