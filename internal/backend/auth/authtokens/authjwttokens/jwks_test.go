package authjwttokens

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWKS(t *testing.T) {
	cfg := &Config{
		Algorithm: AlgorithmRSA,
		RSA: RSAConfig{
			PrivateKey: testRSAPrivateKey,
			PublicKey:  testRSAPublicKey,
		},
	}

	tokens, err := New(cfg)
	require.NoError(t, err)

	rsaTokens, ok := tokens.(RSATokens)
	require.True(t, ok, "tokens should implement RSATokens interface")

	jwks, err := rsaTokens.GetJWKS()
	require.NoError(t, err)
	require.NotNil(t, jwks)
	require.Len(t, jwks.Keys, 1)

	key := jwks.Keys[0]
	assert.Equal(t, "RSA", key.Kty)
	assert.Equal(t, "RS256", key.Alg)
	assert.Equal(t, "sig", key.Use)
	assert.NotEmpty(t, key.N)
	assert.NotEmpty(t, key.E)
}
