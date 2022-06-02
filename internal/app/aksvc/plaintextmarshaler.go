package aksvc

import (
	"errors"
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type plainTextMarshaler struct{}

var PlainTextMarshaler plainTextMarshaler

func (m plainTextMarshaler) Marshal(v interface{}) ([]byte, error) {
	if m, ok := v.(proto.Message); ok {
		return protojson.Marshal(m)
	}

	bs, ok := v.([]byte)
	if !ok {
		return nil, errors.New("not []byte")
	}

	return bs, nil
}

func (m plainTextMarshaler) Unmarshal(data []byte, v interface{}) error {
	if m, ok := v.(proto.Message); ok {
		return protojson.Unmarshal(data, m)
	}

	dst, ok := v.(*[]byte)
	if !ok {
		return errors.New("not *[]byte")
	}

	*dst = make([]byte, len(data))
	copy(*dst, data)

	return nil
}

func (m plainTextMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(func(v interface{}) error {
		bs, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		return m.Unmarshal(bs, v)
	})
}

func (m plainTextMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return runtime.EncoderFunc(func(v interface{}) error {
		bs, err := m.Marshal(v)
		if err != nil {
			return err
		}

		_, err = w.Write(bs)
		return err
	})
}

func (m plainTextMarshaler) ContentType(v interface{}) string {
	return "text/plain"
}
