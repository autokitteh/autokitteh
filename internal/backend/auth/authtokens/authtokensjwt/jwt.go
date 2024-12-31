package authtokensjwt

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
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

type token struct {
	User sdktypes.User
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

func (js *tokens) Create(u sdktypes.User) (string, error) {
	if authusers.IsInternalUserID(u.ID()) {
		return "", sdkerrors.NewInvalidArgumentError("internal user")
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
		Issuer:   issuer,
		Subject:  string(bs),
		ID:       uuid.String(),
	}

	return j.NewWithClaims(method, claim).SignedString(js.signKey)
}

func (js *tokens) Parse(raw string) (sdktypes.User, error) {
	var claims j.RegisteredClaims

	t, err := j.ParseWithClaims(raw, &claims, func(t *j.Token) (interface{}, error) { return js.signKey, nil })
	if err != nil {
		return sdktypes.InvalidUser, err // TODO: better error handling
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
