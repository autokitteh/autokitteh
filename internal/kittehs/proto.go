package kittehs

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func MarshalProtoSliceJSON[T proto.Message](xs []T) ([]byte, error) {
	rs, err := TransformError(xs, func(x T) (json.RawMessage, error) {
		return protojson.Marshal(x)
	})
	if err != nil {
		return nil, err
	}

	return json.Marshal(rs)
}

func MarshalProtoMapJSON[K comparable, V proto.Message](xs map[K]V) ([]byte, error) {
	rs, err := TransformMapValuesError(xs, func(x V) (json.RawMessage, error) {
		return protojson.Marshal(x)
	})
	if err != nil {
		return nil, err
	}

	return json.Marshal(rs)
}
