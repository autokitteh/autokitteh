package sdktypes

const TriggerIDKind = "trg"

type TriggerID = id[triggerIDTraits]

type triggerIDTraits struct{}

func (triggerIDTraits) Prefix() string { return TriggerIDKind }

func NewTriggerID() TriggerID                          { return newID[TriggerID]() }
func ParseTriggerID(s string) (TriggerID, error)       { return ParseID[TriggerID](s) }
func StrictParseTriggerID(s string) (TriggerID, error) { return Strict(ParseTriggerID(s)) }

func IsTriggerID(s string) bool { return IsIDOf[triggerIDTraits](s) }

var InvalidTriggerID TriggerID
