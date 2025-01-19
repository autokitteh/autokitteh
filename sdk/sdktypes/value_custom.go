package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
)

type CustomValuePB = valuev1.Custom

type customValueTraits struct{ immutableObjectTrait }

func (customValueTraits) Validate(m *CustomValuePB) error {
	return errors.Join(
		executorIDField("executor_id", m.ExecutorId),
		objectField[Value]("value", m.Value),
	)
}

func (customValueTraits) StrictValidate(m *CustomValuePB) error {
	return errors.Join(
		mandatory("executor_id", m.ExecutorId),
	)
}

var _ objectTraits[*CustomValuePB] = customValueTraits{}

type CustomValue struct {
	object[*CustomValuePB, customValueTraits]
}

func init() { registerObject[CustomValue]() }

func (CustomValue) isConcreteValue() {}

func (f CustomValue) ExecutorID() ExecutorID {
	return kittehs.Must1(ParseExecutorID(f.read().ExecutorId))
}

func (f CustomValue) Data() []byte { return f.read().Data }

func (f CustomValue) Value() Value { return forceFromProto[Value](f.read().Value) }

func (v Value) IsCustom() bool         { return v.read().Custom != nil }
func (v Value) GetCustom() CustomValue { return forceFromProto[CustomValue](v.read().Custom) }

func init() {
	registerValueGetter(func(v Value) concreteValue {
		if v.IsCustom() {
			return v.GetCustom()
		}
		return nil
	})
}

func NewCustomValue(xid ExecutorID, data []byte, v Value) (Value, error) {
	return ValueFromProto(
		&ValuePB{
			Custom: &CustomValuePB{
				ExecutorId: xid.String(),
				Data:       data,
				Value:      v.ToProto(),
			},
		},
	)
}
