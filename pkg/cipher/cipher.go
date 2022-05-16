package cipher

import (
	"context"
)

type Cipher struct {
	Encrypt func(context.Context, []byte) ([]byte, error)
	Decrypt func(context.Context, []byte) ([]byte, error)
}
