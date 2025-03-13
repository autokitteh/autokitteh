package salesforce

import (
	"context"
	"crypto/tls"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// https://developer.salesforce.com/docs/platform/pub-sub-api/guide/supported-auth.html
// https://pkg.go.dev/google.golang.org/grpc/credentials#PerRPCCredentials
type grpcAuth struct {
	cfg         *oauth2.Config
	token       *oauth2.Token
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
			"MaxAttempts": 4,
			"InitialBackoff": "1s",
			"MaxBackoff": "128s",
			"BackoffMultiplier": 2.0,
			"RetryableStatusCodes": [ "UNAVAILABLE" ]
		}
	}]
}`

func initConn(l *zap.Logger, cfg *oauth2.Config, token *oauth2.Token, instanceURL, orgID string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"api.pubsub.salesforce.com:443",
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&grpcAuth{
			cfg:         cfg,
			token:       token,
			instanceURL: instanceURL,
			tenantID:    orgID,
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
	var err error
	a.token, err = a.cfg.TokenSource(ctx, a.token).Token()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"accesstoken": a.token.AccessToken,
		"instanceurl": a.instanceURL,
		"tenantid":    a.tenantID,
	}, nil
}

func (a *grpcAuth) RequireTransportSecurity() bool {
	return true
}
