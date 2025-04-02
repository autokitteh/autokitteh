package authjwttokens

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	testRSAPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`

	testRSAPublicKey = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEApH5oa4ZuyMKGHLYUeEs0
gAoE+85yOCgP/R1Ma19hB5wd5nrlGya3O45g/3gjhQZMjrJpClOW9hrP5UxFrxQ5
izIqV7/kLr3tCFN+7sNA98BRGKN5KXTVMB8T4YumSRGlUoT3lxXiDfIP49GoLhFB
9NUenuc7/oTxivRUswhJNv6K7Xp3THh7rCH35miKmLRdyaIUTqh1u96JycW2EHLj
GQBwd5BdLRx2AeOEx5V6fXUOUgmPsafI/E8x5BQNvJpvOI9T6/YvuZoG/1BAxB78
kGEJ/vqQeknAnI0IdFza/MVBnTkY0AyKXXSuEQxm83zWnZMK7kact3/ffgO2Ov1f
Q90fxyki4RIN0omZDwSdUYHiJ1+4hbGnFShSNuPZ3jwi5coUZ9//KS3FPm6sP14c
MoNqVph3jY//0eqAze0zqPUCKqHt7taTl+jhMYE1spMQjbtUuTo14OgRRvwF3b6l
mekTFfAgfApT1rBBZCeycY76EtrBkq+8h0c9iuZyYYp1Bd8omCAng2PquWVNIDng
Ccfzd1nsVsjaL+wcnIHs7qo8XrB5Cuy2hvUv7XeJ3YrOsTPDkSkvABQ3MQx80Zx7
y5UeUiwgw1s7ajxc3KXF9yVsesDiO7v+6GHIlM1Go0INHwzUwuL3PxNy1i7cU8pf
5K5Wi7S9b7azLG/f6lGgo5UCAwEAAQ==
-----END PUBLIC KEY-----
`
)

func TestHMACTokens(t *testing.T) {
	cfg := &Config{
		Algorithm: AlgorithmHMAC,
		HMAC: HMACConfig{
			SignKey: strings.Repeat("00", hashSize),
		},
	}

	tokens, err := New(cfg)
	require.NoError(t, err)

	testTokens(t, tokens)
}

func TestRSATokens(t *testing.T) {
	cfg := &Config{
		Algorithm: AlgorithmRSA,
		RSA: RSAConfig{
			PrivateKey: testRSAPrivateKey,
			PublicKey:  testRSAPublicKey,
		},
	}

	tokens, err := New(cfg)
	require.NoError(t, err)

	testTokens(t, tokens)
}

func testTokens(t *testing.T, tokens authtokens.Tokens) {
	t.Helper()

	tests := []struct {
		name    string
		user    sdktypes.User
		wantErr bool
	}{
		{
			name:    "valid user",
			user:    sdktypes.NewUser().WithID(sdktypes.NewUserID()),
			wantErr: false,
		},
		{
			name:    "system user",
			user:    authusers.SystemUser,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tokens.Create(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Test parsing
			parsedUser, err := tokens.Parse(token)
			require.NoError(t, err)
			assert.Equal(t, tt.user.ID(), parsedUser.ID())
		})
	}
}

func TestInvalidTokens(t *testing.T) {
	hmacCfg := &Config{
		Algorithm: AlgorithmHMAC,
		HMAC: HMACConfig{
			SignKey: strings.Repeat("00", hashSize),
		},
	}

	rsaCfg := &Config{
		Algorithm: AlgorithmRSA,
		RSA: RSAConfig{
			PrivateKey: testRSAPrivateKey,
			PublicKey:  testRSAPublicKey,
		},
	}

	tests := []struct {
		name      string
		cfg       *Config
		tokenFunc func(tokens authtokens.Tokens) string
	}{
		{
			name: "invalid hmac token",
			cfg:  hmacCfg,
			tokenFunc: func(tokens authtokens.Tokens) string {
				return "invalid.token.format"
			},
		},
		{
			name: "invalid rsa token",
			cfg:  rsaCfg,
			tokenFunc: func(tokens authtokens.Tokens) string {
				return "invalid.token.format"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := New(tt.cfg)
			require.NoError(t, err)

			invalidToken := tt.tokenFunc(tokens)
			_, err = tokens.Parse(invalidToken)
			assert.Error(t, err)
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid hmac config",
			cfg: &Config{
				Algorithm: AlgorithmHMAC,
				HMAC: HMACConfig{
					SignKey: strings.Repeat("00", hashSize),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid hmac key length",
			cfg: &Config{
				Algorithm: AlgorithmHMAC,
				HMAC: HMACConfig{
					SignKey: "tooshort",
				},
			},
			wantErr: true,
		},
		{
			name: "valid rsa config",
			cfg: &Config{
				Algorithm: AlgorithmRSA,
				RSA: RSAConfig{
					PrivateKey: testRSAPrivateKey,
					PublicKey:  testRSAPublicKey,
				},
			},
			wantErr: false,
		},
		{
			name: "missing rsa private key",
			cfg: &Config{
				Algorithm: AlgorithmRSA,
				RSA: RSAConfig{
					PublicKey: testRSAPublicKey,
				},
			},
			wantErr: true,
		},
		{
			name: "missing rsa public key",
			cfg: &Config{
				Algorithm: AlgorithmRSA,
				RSA: RSAConfig{
					PrivateKey: testRSAPrivateKey,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid algorithm",
			cfg: &Config{
				Algorithm: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCrossAlgorithmTokens(t *testing.T) {
	hmacCfg := &Config{
		Algorithm: AlgorithmHMAC,
		HMAC: HMACConfig{
			SignKey: strings.Repeat("00", hashSize),
		},
	}

	rsaCfg := &Config{
		Algorithm: AlgorithmRSA,
		RSA: RSAConfig{
			PrivateKey: testRSAPrivateKey,
			PublicKey:  testRSAPublicKey,
		},
	}

	hmacTokens, err := New(hmacCfg)
	require.NoError(t, err)

	rsaTokens, err := New(rsaCfg)
	require.NoError(t, err)

	user := sdktypes.NewUser().WithID(sdktypes.NewUserID())

	// Create HMAC token
	hmacToken, err := hmacTokens.Create(user)
	require.NoError(t, err)

	// Create RSA token
	rsaToken, err := rsaTokens.Create(user)
	require.NoError(t, err)

	// Try to parse HMAC token with RSA parser
	_, err = rsaTokens.Parse(hmacToken)
	assert.Error(t, err)

	// Try to parse RSA token with HMAC parser
	_, err = hmacTokens.Parse(rsaToken)
	assert.Error(t, err)
}
