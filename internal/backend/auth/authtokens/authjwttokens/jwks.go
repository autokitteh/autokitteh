package authjwttokens

import (
	"encoding/base64"
	"math/big"
)

// JWKS represents a JSON Web Key Set as defined in RFC 7517
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key as defined in RFC 7517
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// GetJWKS returns the JWKS containing the public key information
func (rs *rsaTokens) GetJWKS() (*JWKS, error) {
	if rs.publicKey == nil {
		return nil, ErrNoPublicKey
	}

	// Convert public key components to base64url encoding
	nBytes := rs.publicKey.N.Bytes()
	eBytes := big.NewInt(int64(rs.publicKey.E)).Bytes()

	jwk := JWK{
		Kid: "1", // You might want to make this configurable or derive it from the key
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		N:   base64.RawURLEncoding.EncodeToString(nBytes),
		E:   base64.RawURLEncoding.EncodeToString(eBytes),
	}

	return &JWKS{
		Keys: []JWK{jwk},
	}, nil
}
