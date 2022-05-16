package kvstore

import (
	"context"
	"encoding/json"
)

func PutJSON(ctx context.Context, s Store, k string, v interface{}) error {
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return s.Put(ctx, k, bs)
}

func GetJSON(ctx context.Context, s Store, k string, v interface{}) error {
	bs, err := s.Get(ctx, k)
	if err != nil {
		return err
	}

	return json.Unmarshal(bs, v)
}
