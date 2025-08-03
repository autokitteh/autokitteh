package authjwttokens

import (
	"errors"
	"time"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const internalIssuer = "internal.autokitteh.cloud"
const internalAudience = "internal.autokitteh.cloud"

type internalTokenData struct {
	j.RegisteredClaims
	Payload map[string]string `json:"payload,omitempty"` // Additional data for internal tokens
}

func createInternalToken(signMethod j.SigningMethod, signKey any, data map[string]string, expiration time.Duration) (string, error) {
	id := uuid.New()
	claims := &internalTokenData{
		RegisteredClaims: j.RegisteredClaims{
			Issuer:    internalIssuer,
			Audience:  []string{internalAudience},
			Subject:   "internal",
			ID:        id.String(),
			ExpiresAt: j.NewNumericDate(time.Now().Add(expiration)),
		},
		Payload: data,
	}

	return j.NewWithClaims(signMethod, claims).SignedString(signKey)
}

func parseInternalToken(algo string, signKey any, raw string) (map[string]string, error) {
	var claims internalTokenData

	t, err := j.ParseWithClaims(raw, &claims, func(t *j.Token) (interface{}, error) {
		if t.Method.Alg() != algo {
			return nil, errors.New("unexpected signing method")
		}
		return signKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !t.Valid {
		return nil, j.ErrSignatureInvalid
	}

	if claims, ok := t.Claims.(*internalTokenData); ok {
		return claims.Payload, nil
	}

	return nil, errors.New("invalid token claims")
}
