package sdktypes

const RunIDKind = "run"

type RunID = *id[runIDTraits]

var _ ID = (RunID)(nil)

type runIDTraits struct{}

func (runIDTraits) Kind() string                   { return RunIDKind }
func (runIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseRunID(raw string) (RunID, error) { return parseTypedID[runIDTraits](raw) }

func StrictParseRunID(raw string) (RunID, error) { return strictParseTypedID[runIDTraits](raw) }

func NewRunID() RunID { return newID[runIDTraits]() }
