package cipher

import (
	"context"
)

var PlaintextCipher = &Cipher{
	Encrypt: func(_ context.Context, data []byte) ([]byte, error) { return data, nil },
	Decrypt: func(_ context.Context, data []byte) ([]byte, error) { return data, nil },
}
