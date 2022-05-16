package apivalues

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	pbvalues "github.com/autokitteh/autokitteh/gen/proto/stubs/go/values"
)

var (
	_ json.Marshaler   = &Value{}
	_ json.Unmarshaler = &Value{}
)

func (v *Value) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(v.pb)
}

func (v *Value) UnmarshalJSON(bs []byte) error {
	if v.pb == nil {
		v.pb = &pbvalues.Value{}
	}

	if err := protojson.Unmarshal(bs, v.pb); err != nil {
		return err
	}

	return v.pb.Validate()
}
