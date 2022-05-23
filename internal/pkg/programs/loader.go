package programs

import (
	"context"
	"errors"

	"go.autokitteh.dev/sdk/api/apiprogram"
)

var ErrNotFound = errors.New("not found")

type LoaderFunc func(context.Context, *apiprogram.Path) ([]byte, error)
