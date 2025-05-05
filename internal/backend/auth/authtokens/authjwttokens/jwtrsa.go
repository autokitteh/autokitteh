package authjwttokens

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	j "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const issuer = "autokitteh.cloud"

var rsaMethod = j.SigningMethodRS256

type RSAConfig struct {
	PrivateKey string `koanf:"private_key"` // PEM encoded RSA private key or file path
	PublicKey  string `koanf:"public_key"`  // PEM encoded RSA public key or file path
}

type rsaTokens struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

var (
	ErrNoPublicKey    = errors.New("no public key available")
	ErrInvalidKeyType = errors.New("invalid key type")
)

type RSATokens interface {
	authtokens.Tokens
	GetJWKS() (*JWKS, error)
}

// isFilePath checks if the input string looks like a file path
func isFilePath(s string) bool {
	return !strings.Contains(s, "-----BEGIN") &&
		(strings.Contains(s, "/") || strings.Contains(s, "\\"))
}

// readFile reads the entire file and returns its contents as a string
func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read file %s: %w", path, err)
	}
	return string(content), nil
}

// getKeyContent returns the key content, reading from a file if necessary
func getKeyContent(key string) (string, error) {
	if isFilePath(key) {
		return readFile(key)
	}
	return key, nil
}

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}

	// Try PKCS1 first
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}

	// If PKCS1 fails, try PKCS8
	pkcs8Key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("key is not an RSA private key")
	}

	return rsaKey, nil
}

func parsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing public key")
	}

	// Try PKIX first
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("key is not an RSA public key")
		}
		return rsaPub, nil
	}

	// If PKIX fails, try x509 certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key is not an RSA public key")
	}

	return rsaPub, nil
}

func newRSA(cfg *RSAConfig) (RSATokens, error) {
	if cfg.PrivateKey == "" || cfg.PublicKey == "" {
		return nil, errors.New("both private and public keys must be provided")
	}

	// Get the private key content (from string or file)
	privateKeyContent, err := getKeyContent(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error reading private key: %w", err)
	}

	// Get the public key content (from string or file)
	publicKeyContent, err := getKeyContent(cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("error reading public key: %w", err)
	}

	// Parse the private key
	privateKey, err := parsePrivateKey(privateKeyContent)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Parse the public key
	publicKey, err := parsePublicKey(publicKeyContent)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	return &rsaTokens{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (rs *rsaTokens) Create(u sdktypes.User) (string, error) {
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
		Issuer:   issuer,
		Subject:  string(bs),
		ID:       uuid.String(),
	}

	return j.NewWithClaims(rsaMethod, claim).SignedString(rs.privateKey)
}

func (rs *rsaTokens) Parse(raw string) (sdktypes.User, error) {
	var claims j.RegisteredClaims

	t, err := j.ParseWithClaims(raw, &claims, func(t *j.Token) (interface{}, error) {
		return rs.publicKey, nil
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
