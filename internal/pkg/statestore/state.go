package statestore

import (
	"context"
	"errors"
	"time"

	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

var ErrNotFound = errors.New("not found")

type Metadata struct {
	UpdatedAt time.Time `json:"updated_at"`
}

type Store interface {
	Set(context.Context, apiproject.ProjectID, string, *apivalues.Value) error
	Get(context.Context, apiproject.ProjectID, string) (*apivalues.Value, *Metadata, error)
	List(context.Context, apiproject.ProjectID) ([]string, error)

	// Will create the variable as List if does not exists and index is 0 or -1.
	Insert(context.Context, apiproject.ProjectID, string, int, *apivalues.Value) error
	Take(_ context.Context, _ apiproject.ProjectID, _ string, idx int, count int) (*apivalues.Value, error)
	Index(context.Context, apiproject.ProjectID, string, int) (*apivalues.Value, error)
	Length(context.Context, apiproject.ProjectID, string) (int, error)

	// Will create the variable as Dict if does not exist.
	SetKey(context.Context, apiproject.ProjectID, string, *apivalues.Value, *apivalues.Value) error

	// Returns nil, nil if key not found.
	GetKey(context.Context, apiproject.ProjectID, string, *apivalues.Value) (*apivalues.Value, error)

	Keys(context.Context, apiproject.ProjectID, string) (*apivalues.Value, error)

	Inc(context.Context, apiproject.ProjectID, string, int64) (*apivalues.Value, error)

	Setup(context.Context) error
	Teardown(context.Context) error
}
