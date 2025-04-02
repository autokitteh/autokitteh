package salesforce

import (
	"context"
	"crypto/tls"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://developer.salesforce.com/docs/platform/pub-sub-api/guide/supported-auth.html
// https://pkg.go.dev/google.golang.org/grpc/credentials#PerRPCCredentials
type grpcAuth struct {
	logger      *zap.Logger
	cfg         *oauth2.Config
	oauth       sdkservices.OAuth
	vars        sdkservices.Vars
	cid         sdktypes.ConnectionID
	instanceURL string
	tenantID    string
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

func (h handler) initConn(cfg *oauth2.Config, cid sdktypes.ConnectionID, instanceURL, orgID string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"api.pubsub.salesforce.com:443",
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&grpcAuth{
			logger:      h.logger,
			cfg:         cfg,
			oauth:       h.oauth,
			vars:        h.vars,
			cid:         cid,
			instanceURL: instanceURL,
			tenantID:    orgID,
		}),
		grpc.WithDefaultServiceConfig(retryPolicy),
	)
	if err != nil {
		h.logger.Error("failed to create gRPC connection for Salesforce events", zap.Error(err))
		return nil, err
	}

	return conn, nil
}

func (a *grpcAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	var err error
	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(a.cid))
	// TODO: add comment. assumes there is a continuous grpc connection that refreshes itself
	// an inactive connection becomes stale after some time.
	token := common.FreshOAuthToken(ctx, a.logger, a.oauth, a.vars, desc, vs)
	if err != nil {
		return nil, err
	}
	// TODO: remove after testing
	a.logger.Warn("refreshed Salesforce OAuth token", zap.String("client_id", a.cid.String()))
	return map[string]string{
		"accesstoken": token.AccessToken,
		"instanceurl": a.instanceURL,
		"tenantid":    a.tenantID,
	}, nil
}

func (a *grpcAuth) RequireTransportSecurity() bool {
	return true
}
