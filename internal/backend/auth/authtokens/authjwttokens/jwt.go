package authjwttokens

import (
	"errors"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Algorithm string

const (
	AlgorithmHMAC Algorithm = "hmac"
	AlgorithmRSA  Algorithm = "rsa"
)

// Config represents the top-level JWT configuration
type Config struct {
	Algorithm Algorithm  `koanf:"algorithm"`
	HMAC      HMACConfig `koanf:"hmac"`
	RSA       RSAConfig  `koanf:"rsa"`
}

// token is shared between HMAC and RSA implementations
type token struct {
	User sdktypes.User
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev: &Config{
		Algorithm: AlgorithmHMAC,
		HMAC: HMACConfig{
			SignKey: strings.Repeat("00", hashSize),
		},
	},
	Test: &Config{
		Algorithm: AlgorithmHMAC,
		HMAC: HMACConfig{
			SignKey: strings.Repeat("00", hashSize),
		},
	},
}

// New creates a new JWT token provider based on the configuration
func New(cfg *Config) (authtokens.Tokens, error) {
	switch cfg.Algorithm {
	case AlgorithmHMAC:
		return newHMAC(&cfg.HMAC)
	case AlgorithmRSA:
		return newRSA(&cfg.RSA)
	default:
		return nil, errors.New("unsupported JWT algorithm")
	}
}
