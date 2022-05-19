package tmplrender

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protojsonMarshalOptions = protojson.MarshalOptions{
	Multiline:     true,
	Indent:        "  ",
	UseProtoNames: true,
}

var funcMap = map[string]interface{}{
	"toDict": func(in interface{}) (interface{}, error) {
		bs, err := json.Marshal(in)
		if err != nil {
			return "", err
		}

		var m map[string]interface{}
		if err := json.Unmarshal(bs, &m); err != nil {
			return nil, err
		}

		return m, nil
	},
	"protoToJson": func(in interface{}) (string, error) {
		bs, err := protojson.Marshal(in.(proto.Message))
		if err != nil {
			return "", err
		}
		return string(bs), nil
	},
	"protoToPrettyJson": func(in interface{}) (string, error) {
		bs, err := protojsonMarshalOptions.Marshal(in.(proto.Message))
		if err != nil {
			return "", err
		}
		return string(bs), nil
	},
	"gets": func(in map[string]string, key string) string {
		return in[key]
	},
}
