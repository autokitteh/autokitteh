package sdktypes

const BuildIDKind = "b"

type BuildID = *id[BuildIDTraits]

var _ ID = (BuildID)(nil)

type BuildIDTraits struct{}

func (BuildIDTraits) Kind() string                   { return BuildIDKind }
func (BuildIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseBuildID(raw string) (BuildID, error) { return parseTypedID[BuildIDTraits](raw) }

func StrictParseBuildID(raw string) (BuildID, error) { return strictParseTypedID[BuildIDTraits](raw) }

func NewBuildID() BuildID { return newID[BuildIDTraits]() }
