package sdktypes

const EnvIDKind = "e"

type EnvID = *id[envIDTraits]

var _ ID = (EnvID)(nil)

type envIDTraits struct{}

func (envIDTraits) Kind() string                   { return EnvIDKind }
func (envIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseEnvID(raw string) (EnvID, error) { return parseTypedID[envIDTraits](raw) }

func StrictParseEnvID(raw string) (EnvID, error) { return strictParseTypedID[envIDTraits](raw) }

func NewEnvID() EnvID { return newID[envIDTraits]() }

func ParseEnvIDOrName(raw string) (Name, EnvID, error) {
	return parseIDOrName[envIDTraits](raw)
}
