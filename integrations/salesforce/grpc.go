package salesforce

import (
	"context"
	"crypto/tls"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://developer.salesforce.com/docs/platform/pub-sub-api/guide/supported-auth.html
// https://pkg.go.dev/google.golang.org/grpc/credentials#PerRPCCredentials
type grpcAuth struct {
	logger *zap.Logger
	vars   sdkservices.Vars
	oauth  *oauth.OAuth
	vsid   sdktypes.VarScopeID
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

func (h handler) initConn(l *zap.Logger, cid sdktypes.ConnectionID) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"api.pubsub.salesforce.com:443",
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&grpcAuth{
			logger: l,
			vars:   h.vars,
			oauth:  h.oauth,
			vsid:   sdktypes.NewVarScopeID(cid),
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
	vs, err := a.vars.Get(authcontext.SetAuthnSystemUser(ctx), a.vsid)
	if err != nil {
		a.logger.Error("failed to read connection vars", zap.Error(err))
		return nil, err
	}

	t := a.oauth.FreshToken(ctx, a.logger, desc, vs)

	return map[string]string{
		"accesstoken": t.AccessToken,
		"instanceurl": vs.GetValue(instanceURLVar),
		"tenantid":    vs.GetValue(orgIDVar),
	}, nil
}

func (a *grpcAuth) RequireTransportSecurity() bool {
	return true
}
