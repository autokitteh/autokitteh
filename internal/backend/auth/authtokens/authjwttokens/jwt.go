package authjwttokens

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"crypto/ecdsa"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	issuer       = "https://api.autokitteh.cloud"
	oneYear      = 365 * 24 * time.Hour
	devPublicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEpn+ugYKvxyjH1PP5M+TyI9AhxnWP
3NtBzkt35Ppv9aX2YoBMXbbQcgUFDCwKY3QpsCmUILh3vno97lHkbgwjbQ==
-----END PUBLIC KEY-----`
	devPrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBB5QU93Ceqzqc1Yn415+nDBkTQ7Zdrs1rxV2bvNIXlQoAoGCCqGSM49
AwEHoUQDQgAEpn+ugYKvxyjH1PP5M+TyI9AhxnWP3NtBzkt35Ppv9aX2YoBMXbbQ
cgUFDCwKY3QpsCmUILh3vno97lHkbgwjbQ==
-----END EC PRIVATE KEY-----
`
)

var (
	method   = j.SigningMethodES256
	hashSize = method.Hash.Size()
)

type Config struct {
	PrivateKey            string        `koanf:"private_key"`
	PublicKey             string        `koanf:"public_key"`
	ExpirationTimeMinutes time.Duration `koanf:"expiration_time_minutes"`
}

type tokens struct {
	publicKey             *ecdsa.PublicKey
	privateKey            *ecdsa.PrivateKey
	expirationTimeMinutes time.Duration
}

type token struct {
	User sdktypes.User
}

var Configs = configset.Set[Config]{
	Default: &Config{
		ExpirationTimeMinutes: oneYear,
	},
	Dev: &Config{
		PublicKey:  devPublicKey,
		PrivateKey: devPrivateKey,
	},
	Test: &Config{
		PublicKey:  devPublicKey,
		PrivateKey: devPrivateKey,
	},
}

func New(cfg *Config) (authtokens.Tokens, error) {
	privateKey, err := j.ParseECPrivateKeyFromPEM([]byte(cfg.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("invalid signing key: %w", err)
	}

	publicKey, err := j.ParseECPublicKeyFromPEM([]byte(cfg.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("invalid signing key: %w", err)
	}

	return &tokens{privateKey: privateKey, publicKey: publicKey}, nil
}

func (js *tokens) Create(u sdktypes.User) (string, error) {
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
		IssuedAt:  j.NewNumericDate(time.Now()),
		Audience:  j.ClaimStrings{u.ID().String()},
		Issuer:    issuer,
		Subject:   string(bs),
		ID:        uuid.String(),
		ExpiresAt: j.NewNumericDate(time.Now().Add(time.Minute * js.expirationTimeMinutes)),
	}

	return j.NewWithClaims(method, claim).SignedString(js.privateKey)
}

func (js *tokens) Parse(raw string) (sdktypes.User, error) {
	var claims j.RegisteredClaims

	t, err := j.ParseWithClaims(raw, &claims, func(t *j.Token) (interface{}, error) { return js.publicKey, nil })
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
