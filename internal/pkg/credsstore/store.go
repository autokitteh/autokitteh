package credsstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/pkg/pkvstore"
)

var ErrNotFound = pkvstore.ErrNotFound

const anyPID = apiproject.ProjectID("any")

type Store struct{ Store pkvstore.Store }

type record struct {
	Memo  interface{} `json:"memo,omitempty"`
	Value []byte      `json:"value"`
}

func key(k, n string) string { return fmt.Sprintf("%s.%s", k, n) }

func (s *Store) Set(ctx context.Context, pid apiproject.ProjectID, kind, name string, v []byte, memo interface{}) error {
	if pid.IsEmpty() {
		pid = anyPID
	}

	bs, err := json.Marshal(record{Memo: memo, Value: v})
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return s.Store.Put(ctx, pid.String(), key(kind, name), bs)
}

func (s *Store) Get(ctx context.Context, pid apiproject.ProjectID, kind, name string) ([]byte, error) {
	if pid.IsEmpty() {
		pid = anyPID
	}

	v, err := s.Store.Get(ctx, pid.String(), key(kind, name))
	if err != nil && pid != anyPID && errors.Is(err, pkvstore.ErrNotFound) {
		v, err = s.Store.Get(ctx, anyPID.String(), key(kind, name))
	}

	var r record
	if err := json.Unmarshal(v, &r); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return r.Value, err
}

func (s *Store) Setup(ctx context.Context) error { return s.Store.Setup(ctx) }
