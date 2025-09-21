package authjwttokens

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// We might update this to api.autokitteh.cloud
// in the future
const issuerBase = "autokitteh.cloud"

type externalTokenData struct {
	j.RegisteredClaims
	Payload map[string]string `json:"payload,omitempty"` // Additional data for internal tokens
}

func createExternalToken(signMethod j.SigningMethod, signKey any, data []byte, internalUser bool) (string, error) {
	id := uuid.New()
	aud := []string{"api." + issuerBase}
	if internalUser {
		aud = append(aud, "internal."+issuerBase)
	}

	claims := &externalTokenData{
		RegisteredClaims: j.RegisteredClaims{
			Issuer:   issuerBase,
			Audience: aud,
			Subject:  string(data),
			ID:       id.String(),
			IssuedAt: j.NewNumericDate(time.Now()),
		},
	}

	return j.NewWithClaims(signMethod, claims).SignedString(signKey)
}

func parseExternalToken(algo string, signKey any, raw string) (sdktypes.User, error) {
	var claims externalTokenData

	t, err := j.ParseWithClaims(raw, &claims, func(t *j.Token) (interface{}, error) {
		if t.Method.Alg() != algo {
			return nil, errors.New("unexpected signing method")
		}
		return signKey, nil
	})
	if err != nil {
		return sdktypes.InvalidUser, err
	}

	if !t.Valid {
		return sdktypes.InvalidUser, j.ErrSignatureInvalid
	}

	var tok token
	if err := json.Unmarshal([]byte(claims.Subject), &tok); err != nil {
		return sdktypes.InvalidUser, fmt.Errorf("unmarshal token: %w", err)
	}

	return tok.User, nil

}
