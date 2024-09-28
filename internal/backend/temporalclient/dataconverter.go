package temporalclient

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"go.temporal.io/sdk/converter"
)

type DataConverterEncryptionConfig struct {
	// If `Encrypt` is true, `KeyNames` must have at least one key name specified.
	Encrypt bool `koanf:"encrypt"`

	// Comma-separated list of key names and values.
	// First key is used for encryption, others are used only for decryption.
	// Example: "key1=<64 char hex>,keys=<64 char hex>"
	Keys string `koanf:"keys"`
}

type DataConverterConfig struct {
	Compress   bool                          `koanf:"compress"`
	Encryption DataConverterEncryptionConfig `koanf:"encryption"`
}

var (
	ErrNoKeys      = errors.New("at least one encryption key name must be provided")
	ErrKeyNotFound = errors.New("encryption key not found in environment")
)

func NewDataConverter(cfg *DataConverterConfig, parent converter.DataConverter) (converter.DataConverter, error) {
	var codecs []converter.PayloadCodec

	if cfg.Encryption.Encrypt {
		if len(cfg.Encryption.Keys) == 0 {
			return nil, ErrNoKeys
		}

		pairs := strings.Split(cfg.Encryption.Keys, ",")

		codec := encryptionCodec{
			ciphers: make(map[string]cipher.AEAD, len(pairs)),
		}

		for _, pair := range pairs {
			n, v, ok := strings.Cut(pair, "=")
			if !ok {
				return nil, fmt.Errorf("invalid key-value pair %q", pair)
			}

			n, v = strings.TrimSpace(n), strings.TrimSpace(v)

			if len(n) == 0 {
				return nil, fmt.Errorf("empty key name")
			}

			if codec.main == "" {
				codec.main = n
			}

			var err error
			if codec.ciphers[n], err = newCipher(v); err != nil {
				return nil, fmt.Errorf("key %q: %w", n, err)
			}
		}

		codecs = append(codecs, &codec)
	}

	if cfg.Compress {
		codecs = append(codecs, converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}))
	}

	return converter.NewCodecDataConverter(parent, codecs...), nil
}

func newCipher(hexKey string) (cipher.AEAD, error) {
	bs, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: %w", err)
	}

	c, err := aes.NewCipher(bs)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	return gcm, nil
}
