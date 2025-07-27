package internalclient

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

type cli struct {
	tokens           authtokens.Tokens
	l                *zap.Logger
	internalEndpoint string
}

type InternalClient interface {
	NewOrgImpersonator(orgID sdktypes.OrgID) (sdkservices.Services, error)
}

func New(tokens authtokens.Tokens, l *zap.Logger, cfg *Config) (InternalClient, error) {
	if cfg.InternalEndpoint == "" {
		return nil, fmt.Errorf("internal endpoint is not configured")
	}

	return &cli{
		tokens:           tokens,
		l:                l,
		internalEndpoint: cfg.InternalEndpoint,
	}, nil
}

func (c *cli) NewOrgImpersonator(orgID sdktypes.OrgID) (sdkservices.Services, error) {
	internalToken, err := c.tokens.CreateInternal(map[string]string{
		"orgID": orgID.UUIDValue().String(),
	})

	if err != nil {
		return nil, fmt.Errorf("create internal token: %w", err)
	}

	cli := sdkclients.New(sdkclient.Params{
		URL:       c.internalEndpoint,
		AuthToken: internalToken,
	}.Safe())

	c.l.Debug("created internal client for org", zap.String("orgID", orgID.UUIDValue().String()))
	return cli, nil
}
