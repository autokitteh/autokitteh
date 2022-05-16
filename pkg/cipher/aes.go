package cipher

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func NewAESCipher(key []byte) (*Cipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()

	return &Cipher{
		Encrypt: func(_ context.Context, data []byte) ([]byte, error) {
			nonce := make([]byte, 12)
			if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
				return nil, err
			}

			return gcm.Seal(nonce, nonce, data, nil), nil
		},
		Decrypt: func(_ context.Context, data []byte) ([]byte, error) {
			nonce, encrypted := data[:nonceSize], data[nonceSize:]

			decrypted, err := gcm.Open(nil, nonce, encrypted, nil)
			if err != nil {
				return nil, err
			}

			return decrypted, nil
		},
	}, nil
}
