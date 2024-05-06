package kittehs

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
)

func EncodeURLData(data any) (string, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(data); err != nil {
		return "", err
	}

	var z bytes.Buffer
	w := gzip.NewWriter(&z)
	if _, err := w.Write(b.Bytes()); err != nil {
		return "", err
	}

	if err := w.Close(); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(z.Bytes()), nil
}

func DecodeURLData[T any](data string, dst T) error {
	b, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return err
	}

	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}

	if err := json.NewDecoder(r).Decode(&dst); err != nil {
		return err
	}

	return nil
}
