package authjwttokens

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const hmacIssuer = "ak-hmac"

var (
	hmacMethod = j.SigningMethodHS256
	hashSize   = hmacMethod.Hash.Size()
)

type HMACConfig struct {
	SignKey string `koanf:"sign_key"`
}

type hmacTokens struct {
	signKey []byte
}

func newHMAC(cfg *HMACConfig) (authtokens.Tokens, error) {
	key, err := hex.DecodeString(cfg.SignKey)
	if err != nil {
		return nil, fmt.Errorf("invalid signing key: %w", err)
	}

	if len(key) != hashSize {
		return nil, fmt.Errorf("invalid key len: %d != desired %d", len(key), hashSize)
	}

	return &hmacTokens{signKey: key}, nil
}

func (js *hmacTokens) Create(u sdktypes.User) (string, error) {
	if authusers.IsSystemUserID(u.ID()) {
		return "", sdkerrors.NewInvalidArgumentError("system user")
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate UUID: %w", err)
	}

	tok := token{User: u}
	bs, err := json.Marshal(tok)
	if err != nil {
		return "", fmt.Errorf("marshal token: %w", err)
	}

	claim := j.RegisteredClaims{
		IssuedAt: j.NewNumericDate(time.Now()),
		Issuer:   hmacIssuer,
		Subject:  string(bs),
		ID:       uuid.String(),
	}

	return j.NewWithClaims(hmacMethod, claim).SignedString(js.signKey)
}

func (js *hmacTokens) Parse(raw string) (sdktypes.User, error) {
	var claims j.RegisteredClaims

	t, err := j.ParseWithClaims(raw, &claims, func(t *j.Token) (interface{}, error) {
		return js.signKey, nil
	})
	if err != nil {
		return sdktypes.InvalidUser, err
	}

	if !t.Valid {
		return sdktypes.InvalidUser, errors.New("invalid token")
	}

	var tok token
	if err := json.Unmarshal([]byte(claims.Subject), &tok); err != nil {
		return sdktypes.InvalidUser, fmt.Errorf("unmarshal token: %w", err)
	}

	return tok.User, nil
}
