package authjwttokens

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	j "github.com/golang-jwt/jwt/v5"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

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

	tok := token{User: u}
	bs, err := json.Marshal(tok)
	if err != nil {
		return "", fmt.Errorf("marshal token: %w", err)
	}
	return createExternalToken(hmacMethod, js.signKey, bs, true)
}

func (js *hmacTokens) Parse(raw string) (sdktypes.User, error) {
	return parseExternalToken(hmacMethod.Alg(), js.signKey, raw)
}

func (js *hmacTokens) CreateInternal(data map[string]string) (string, error) {
	return createInternalToken(hmacMethod, js.signKey, data, 10*time.Minute)
}

func (js *hmacTokens) ParseInternal(raw string) (map[string]string, error) {
	return parseInternalToken(hmacMethod.Alg(), js.signKey, raw)
}
