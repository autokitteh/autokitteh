package sdktypes

const MappingIDKind = "m"

type MappingID = *id[mappingIDTraits]

var _ ID = (MappingID)(nil)

type mappingIDTraits struct{}

func (mappingIDTraits) Kind() string                   { return MappingIDKind }
func (mappingIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseMappingID(raw string) (MappingID, error) { return parseTypedID[mappingIDTraits](raw) }

func StrictParseMappingID(raw string) (MappingID, error) {
	return strictParseTypedID[mappingIDTraits](raw)
}

func NewMappingID() MappingID { return newID[mappingIDTraits]() }
