package salesforce

import (
	"context"
	"crypto/tls"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type salesforceAuth struct {
	accessToken string
	instanceURL string
	tenantID    string
}

func initConn(accessToken, instanceURL, orgID string, l *zap.Logger) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		"api.pubsub.salesforce.com:443",
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(&salesforceAuth{
			accessToken: accessToken,
			instanceURL: instanceURL,
			tenantID:    orgID,
		}),
	)
	if err != nil {
		l.Error("failed to create gRPC connection", zap.Error(err))
		return nil, err
	}
	return conn, nil
}

func (a *salesforceAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"accesstoken": a.accessToken,
		"instanceurl": a.instanceURL,
		"tenantid":    a.tenantID,
	}, nil
}

func (a *salesforceAuth) RequireTransportSecurity() bool {
	return true
}
