package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/descope/go-sdk/descope/client"
	"github.com/descope/go-sdk/descope/logger"
	"go.uber.org/zap"
)

type descopeAuthenticator struct {
	cfg  *Config
	z    *zap.Logger
	c    *client.DescopeClient
	name string
}

func stringToLogLevel(level string) logger.LogLevel {
	switch level {
	case "none":
		return logger.LogNone
	case "debug":
		return logger.LogDebugLevel
	}

	return logger.LogInfoLevel
}

func newDescopeAuthenticator(cfg *Config, z *zap.Logger) (Authenticator, error) {
	if cfg.ProjectID == "" {
		return nil, ErrInvalidAuthProviderConfiguration
	}

	ll := stringToLogLevel(cfg.LogLevel)

	descopeClient, err := client.NewWithConfig(&client.Config{ProjectID: cfg.ProjectID, LogLevel: ll})
	if err != nil {
		z.Error("failed initiating descope: %w", zap.Error(err))
		return nil, ErrInvalidAuthProviderConfiguration
	}

	return &descopeAuthenticator{
		cfg:  cfg,
		z:    z,
		c:    descopeClient,
		name: AuthProviderDescope,
	}, nil
}

func (d *descopeAuthenticator) validateToken(req *http.Request) (bool, *AuthenticatedUserDetails) {
	authorized, tok, err := d.c.Auth.ValidateSessionWithRequest(req)
	if !authorized {
		if err != nil {
			d.z.Error("validate session error", zap.Error(err))
		}
		return false, nil
	}

	return true, &AuthenticatedUserDetails{UserID: tok.ID, Provider: d.name}
}

func (d *descopeAuthenticator) Middleware(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorized, ud := d.validateToken(r)
		if !authorized {
			w.WriteHeader(401)
			res := map[string]any{"error": "unauthorized"}
			json.NewEncoder(w).Encode(res)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, AuthenticatedUserCtxKey, ud)

		n.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (d *descopeAuthenticator) Provider() ProviderDetails {
	return ProviderDetails{Name: AuthProviderDescope, Config: map[string]string{"ProjectID": d.cfg.ProjectID}}
}
