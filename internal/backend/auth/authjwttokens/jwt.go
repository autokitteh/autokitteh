package authjwttokens

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const issuer = "ak"

var (
	method   = j.SigningMethodHS256
	hashSize = method.Hash.Size()
)

type Config struct {
	SignKey string `koanf:"sign_key"`
}

type tokens struct {
	signKey []byte
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev: &Config{
		SignKey: strings.Repeat("00", hashSize),
	},
	Test: &Config{
		SignKey: strings.Repeat("00", hashSize),
	},
}

func New(cfg *Config) (authtokens.Tokens, error) {
	key, err := hex.DecodeString(cfg.SignKey)
	if err != nil {
		return nil, fmt.Errorf("invalid signing key: %w", err)
	}

	if len(key) != hashSize {
		return nil, fmt.Errorf("invalid key len: %d != desired %d", len(key), hashSize)
	}

	return &tokens{signKey: key}, nil
}

func (js *tokens) Create(uid sdktypes.UserID) (string, error) {
	uj, err := uid.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate UUID: %w", err)
	}

	claim := j.RegisteredClaims{
		IssuedAt: j.NewNumericDate(time.Now()),
		Issuer:   issuer,
		Subject:  string(uj),
		ID:       uuid.String(),
	}

	return j.NewWithClaims(method, claim).SignedString(js.signKey)
}

func (js *tokens) Parse(token string) (sdktypes.UserID, error) {
	var claims j.RegisteredClaims

	t, err := j.ParseWithClaims(token, &claims, func(t *j.Token) (interface{}, error) { return js.signKey, nil })
	if err != nil {
		return sdktypes.InvalidUserID, err // TODO: better error handling
	}

	if !t.Valid {
		return sdktypes.InvalidUserID, errors.New("invalid token")
	}

	var uid sdktypes.UserID
	if err := uid.UnmarshalJSON([]byte(claims.Subject)); err != nil {
		return sdktypes.InvalidUserID, fmt.Errorf("unmarshal JSON: %w", err)
	}

	return uid, nil
}
