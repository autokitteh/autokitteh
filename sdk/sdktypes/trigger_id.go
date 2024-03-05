package sdktypes

const triggerIDKind = "trg"

type TriggerID = id[triggerIDTraits]

type triggerIDTraits struct{}

func (triggerIDTraits) Prefix() string { return triggerIDKind }

func NewTriggerID() TriggerID                          { return newID[TriggerID]() }
func ParseTriggerID(s string) (TriggerID, error)       { return ParseID[TriggerID](s) }
func StrictParseTriggerID(s string) (TriggerID, error) { return Strict(ParseTriggerID(s)) }

var InvalidTriggerID TriggerID
