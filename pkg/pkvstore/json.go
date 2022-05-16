package pkvstore

import (
	"context"
	"encoding/json"
)

func PutJSON(ctx context.Context, s Store, p, k string, v interface{}) error {
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return s.Put(ctx, p, k, bs)
}

func GetJSON(ctx context.Context, s Store, p, k string, v interface{}) error {
	bs, err := s.Get(ctx, p, k)
	if err != nil {
		return err
	}

	return json.Unmarshal(bs, v)
}
