package auth

import (
	"net/http"

	"go.uber.org/zap"
)

type ProviderDetails struct {
	Name   string
	Config map[string]string
}

type Authenticator interface {
	Middleware(next http.Handler) http.Handler
	Provider() ProviderDetails
}

const AuthenticatedUserCtxKey = "auth:user_details"

type AuthenticatedUserDetails struct {
	UserID   string
	Provider string
}

var (
	DefaultUser = AuthenticatedUserDetails{
		UserID:   "default",
		Provider: "none",
	}
)

const (
	AuthProviderNone    = "none"
	AuthProviderDescope = "descope"
)

func NewAuthenticator(cfg *Config, z *zap.Logger) (Authenticator, error) {
	if !cfg.Enabled {
		z.Info("authentication is disabled")
		return &noAuthAuthenticator{}, nil
	}

	switch cfg.Provider {
	case "descope":
		return newDescopeAuthenticator(cfg, z)
	}

	return nil, ErrNotSupportedAuthProvider
}
