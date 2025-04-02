package salesforce

import (
	"context"
	"crypto/tls"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://developer.salesforce.com/docs/platform/pub-sub-api/guide/supported-auth.html
// https://pkg.go.dev/google.golang.org/grpc/credentials#PerRPCCredentials
type grpcAuth struct {
	cfg         *oauth2.Config
	token       *oauth2.Token
	instanceURL string
	tenantID    string
	oauth       *oauth.OAuth
	logger      *zap.Logger
	integration sdktypes.Integration
	vars        sdktypes.Vars
}

// Based on:
// - https://grpc.io/docs/guides/retry/
// - https://github.com/grpc/grpc-proto/blob/master/grpc/service_config/service_config.proto
const retryPolicy = `{
	"methodConfig": [{
		"name": [{"service": "eventbus.v1.PubSub", "method": "Subscribe"}],
		"retryPolicy": {
			"maxAttempts": 4,
			"initialBackoff": "1s",
			"maxBackoff": "10s",
			"backoffMultiplier": 2.0,
			"retryableStatusCodes": [
				"CANCELLED", 
				"UNKNOWN", 
				"INVALID_ARGUMENT",
				"DEADLINE_EXCEEDED",
				"NOT_FOUND",
				"ALREADY_EXISTS",
				"PERMISSION_DENIED",
				"RESOURCE_EXHAUSTED",
				"FAILED_PRECONDITION",
				"ABORTED",
				"OUT_OF_RANGE",
				"UNIMPLEMENTED",
				"INTERNAL",
				"UNAVAILABLE",
				"DATA_LOSS",
				"UNAUTHENTICATED"
			]
		}
	}]
}`

func initConn(l *zap.Logger, cfg *oauth2.Config, token *oauth2.Token, instanceURL, orgID string, oauth *oauth.OAuth, integration sdktypes.Integration, vars sdktypes.Vars) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"api.pubsub.salesforce.com:443",
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&grpcAuth{
			cfg:         cfg,
			token:       token,
			instanceURL: instanceURL,
			tenantID:    orgID,
			oauth:       oauth,
			logger:      l,
			integration: integration,
			vars:        vars,
		}),
		grpc.WithDefaultServiceConfig(retryPolicy),
	)
	if err != nil {
		l.Error("failed to create gRPC connection for Salesforce events", zap.Error(err))
		return nil, err
	}

	return conn, nil
}

func (a *grpcAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	a.token = a.oauth.FreshToken(ctx, a.logger, a.integration, a.vars)

	return map[string]string{
		"accesstoken": a.token.AccessToken,
		"instanceurl": a.instanceURL,
		"tenantid":    a.tenantID,
	}, nil
}

func (a *grpcAuth) RequireTransportSecurity() bool {
	return true
}
