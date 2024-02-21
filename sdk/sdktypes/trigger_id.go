package sdktypes

const TriggerIDKind = "t"

type TriggerID = *id[triggerIDTraits]

var _ ID = (TriggerID)(nil)

type triggerIDTraits struct{}

func (triggerIDTraits) Kind() string                   { return TriggerIDKind }
func (triggerIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseTriggerID(raw string) (TriggerID, error) { return parseTypedID[triggerIDTraits](raw) }

func StrictParseTriggerID(raw string) (TriggerID, error) {
	return strictParseTypedID[triggerIDTraits](raw)
}

func NewTriggerID() TriggerID { return newID[triggerIDTraits]() }
