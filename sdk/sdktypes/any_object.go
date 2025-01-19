package sdktypes

import (
	"encoding/json"
	"fmt"
)

// Used to serialize objects as polymorphic JSON objects.
type AnyObject struct{ Object }

func NewAnyObject(o Object) AnyObject    { return AnyObject{o} }
func UnwrapAnyObject(o AnyObject) Object { return o.Object }

type anyObject struct {
	Kind   string          `json:"kind"`
	Object json.RawMessage `json:"object"`
}

var msgs = make(map[string]func([]byte) (Object, error))

type registeredObject[M comparableMessage, T objectTraits[M]] interface {
	Object
	~struct{ object[M, T] }
}

func registerObject[O registeredObject[M, T], M comparableMessage, T objectTraits[M]]() {
	var o O
	msgs[o.ProtoMessageName()] = func(bs []byte) (Object, error) {
		o, err := o.NewFromJSON(bs)
		if err != nil {
			return nil, err
		}

		// The result object need to be typed according to the original registration, not
		// the underlying embedded object[M, T] type.
		return O{o.(object[M, T])}, nil
	}
}

func (a AnyObject) MarshalJSON() ([]byte, error) {
	// The object can be marshalled as a pointer, so if it's null, we just
	// specify null in JSON.
	if !a.IsValid() {
		return []byte("null"), nil
	}

	bs, err := a.Object.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return json.Marshal(anyObject{
		Kind:   a.ProtoMessageName(),
		Object: bs,
	})
}

func (a *AnyObject) UnmarshalJSON(b []byte) error {
	var raw anyObject

	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	unmarshal, ok := msgs[raw.Kind]
	if !ok {
		return fmt.Errorf("unknown kind %q", raw.Kind)
	}

	obj, err := unmarshal(raw.Object)
	if err != nil {
		return err
	}

	a.Object = obj
	return nil
}
