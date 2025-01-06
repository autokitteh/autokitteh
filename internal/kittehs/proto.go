package kittehs

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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

// ProtoToMap converts a protobuf message to a map, using the provided field mask.
// If fm is nil, all fields are included.
// This function deals only with top level fields, nested fields are not supported.
// IMPORTANT: This currently supports only scalar types as values.
func ProtoToMap(pb proto.Message, fm *fieldmaskpb.FieldMask) (map[string]interface{}, error) {
	v := pb.ProtoReflect()
	fs := v.Descriptor().Fields()
	m := make(map[string]any, fs.Len())

	if fm == nil {
		for i := range fs.Len() {
			fd := fs.Get(i)
			m[string(fd.Name())] = v.Get(fd).Interface()
		}

		return m, nil
	}

	if !fm.IsValid(pb) {
		return nil, fmt.Errorf("invalid field mask for %T", pb)
	}

	for _, p := range fm.Paths {
		fd := fs.ByName(protoreflect.Name(p))
		if fd == nil {
			return nil, fmt.Errorf("field %q not found in %T", p, pb)
		}

		fv := v.Get(fd)
		av := fv.Interface()

		switch fd.Kind() {
		case protoreflect.EnumKind:
			av = int(fv.Enum())
		case protoreflect.MessageKind, protoreflect.GroupKind:
			return nil, fmt.Errorf("unsupported field kind %v", fd.Kind())
		}

		m[p] = av
	}

	return m, nil
}
