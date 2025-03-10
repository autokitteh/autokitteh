package salesforce

import (
	"context"
	"crypto/tls"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// https://developer.salesforce.com/docs/platform/pub-sub-api/guide/supported-auth.html
// https://pkg.go.dev/google.golang.org/grpc/credentials#PerRPCCredentials
type grpcAuth struct {
	accessToken string
	instanceURL string
	tenantID    string
}

func initConn(l *zap.Logger, accessToken, instanceURL, orgID string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"api.pubsub.salesforce.com:443",
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&grpcAuth{
			accessToken: accessToken,
			instanceURL: instanceURL,
			tenantID:    orgID,
		}),
	)
	if err != nil {
		l.Error("failed to create gRPC connection for Salesforce events", zap.Error(err))
		return nil, err
	}
	return conn, nil
}

func (a *grpcAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"accesstoken": a.accessToken,
		"instanceurl": a.instanceURL,
		"tenantid":    a.tenantID,
	}, nil
}

func (a *grpcAuth) RequireTransportSecurity() bool {
	return true
}
