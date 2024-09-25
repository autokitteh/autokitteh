package temporalclient

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.temporal.io/sdk/converter"
)

const (
	configKeyEnvVarPrefix = "AK_DATACONV_ENCRYPTION_KEY_"
)

type DataConverterEncryptionConfig struct {
	// If `Encrypt` is true, `KeyNames` must have at least one key name specified.
	Encrypt bool `koanf:"encrypt"`

	// Comma-separated list of key names.
	// First key in the active encryption key. All the rest are for decryption only.
	// For each name, the encryption key is fetched from the environment, named as
	// "AK_DATACONV_ENCRYPTION_KEY_<name>". Each key is 64 characters long, hex-encoded.
	KeyNames string `koanf:"key_names"`
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
		if len(cfg.Encryption.KeyNames) == 0 {
			return nil, ErrNoKeys
		}

		names := strings.Split(cfg.Encryption.KeyNames, ",")

		ciphers := make(map[string]cipher.AEAD, len(names))
		for _, name := range names {
			var err error
			if ciphers[name], err = newCipher(strings.TrimSpace(name)); err != nil {
				return nil, err
			}
		}

		codecs = append(codecs, &encryptionCodec{main: names[0], ciphers: ciphers})
	}

	if cfg.Compress {
		codecs = append(codecs, converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}))
	}

	return converter.NewCodecDataConverter(parent, codecs...), nil
}

func newCipher(name string) (cipher.AEAD, error) {
	key, ok := os.LookupEnv(configKeyEnvVarPrefix + strings.ToUpper(name))
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrKeyNotFound, name)
	}

	bs, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key %q: %w", name, err)
	}

	c, err := aes.NewCipher(bs)
	if err != nil {
		return nil, fmt.Errorf("new cipher %q: %w", name, err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, fmt.Errorf("new gcm %q: %w", name, err)
	}

	return gcm, nil
}
