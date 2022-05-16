package programs

import (
	"context"
	"errors"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiprogram"
)

var ErrNotFound = errors.New("not found")

type LoaderFunc func(context.Context, *apiprogram.Path) ([]byte, error)
