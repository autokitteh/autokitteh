package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	triggersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1"
)

type triggerSourceTypeTraits struct{}

var _ enumTraits = triggerSourceTypeTraits{}

func (triggerSourceTypeTraits) Prefix() string           { return "SOURCE_TYPE_" }
func (triggerSourceTypeTraits) Names() map[int32]string  { return triggersv1.Trigger_SourceType_name }
func (triggerSourceTypeTraits) Values() map[string]int32 { return triggersv1.Trigger_SourceType_value }

type TriggerSourceType struct {
	enum[triggerSourceTypeTraits, triggersv1.Trigger_SourceType]
}

func triggerStateFromProto(e triggersv1.Trigger_SourceType) TriggerSourceType {
	return kittehs.Must1(TriggerSourceTypeFromProto(e))
}

var (
	PossibleTriggerSourceTypesNames = AllEnumNames[triggerSourceTypeTraits]()

	TriggerSourceTypeUnspecified = triggerStateFromProto(triggersv1.Trigger_SOURCE_TYPE_UNSPECIFIED)
	TriggerSourceTypeConnection  = triggerStateFromProto(triggersv1.Trigger_SOURCE_TYPE_CONNECTION)
	TriggerSourceTypeWebhook     = triggerStateFromProto(triggersv1.Trigger_SOURCE_TYPE_WEBHOOK)
	TriggerSourceTypeSchedule    = triggerStateFromProto(triggersv1.Trigger_SOURCE_TYPE_SCHEDULE)
)

func TriggerSourceTypeFromProto(e triggersv1.Trigger_SourceType) (TriggerSourceType, error) {
	return EnumFromProto[TriggerSourceType](e)
}

func ParseTriggerSourceType(raw string) (TriggerSourceType, error) {
	return ParseEnum[TriggerSourceType](raw)
}
