package sdktypes

const MappingEntrypointIDKind = "meid"

type MappingEntrypointID = *id[mappingEntrypointIDTraits]

var _ ID = (MappingEntrypointID)(nil)

type mappingEntrypointIDTraits struct{}

func (mappingEntrypointIDTraits) Kind() string                   { return MappingEntrypointIDKind }
func (mappingEntrypointIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseMappingEntrypointID(raw string) (MappingEntrypointID, error) {
	return parseTypedID[mappingEntrypointIDTraits](raw)
}

func StrictParseMappingEntrypointID(raw string) (MappingEntrypointID, error) {
	return strictParseTypedID[mappingEntrypointIDTraits](raw)
}

func NewEMappingEntrypointD() MappingEntrypointID { return newID[mappingEntrypointIDTraits]() }
