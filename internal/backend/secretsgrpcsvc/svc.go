package secretsgrpcsvc

import (
	"context"

	"connectrpc.com/connect"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	secretsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/secrets/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/secrets/v1/secretsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type server struct {
	impl   sdkservices.Secrets
	logger *zap.Logger

	secretsv1connect.UnimplementedSecretsServiceHandler
}

var _ secretsv1connect.SecretsServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, l *zap.Logger, sec sdkservices.Secrets) error {
	s := server{impl: sec, logger: l}
	path, handler := secretsv1connect.NewSecretsServiceHandler(&s)
	muxes.Auth.Handle(path, handler)
	return nil
}

// Create generates a new token to represent a connection's specified
// key-value data, and associates them bidirectionally. If the same
// request is sent N times, this method returns N different tokens.
func (s server) Create(ctx context.Context, req *connect.Request[secretsv1.CreateRequest]) (*connect.Response[secretsv1.CreateResponse], error) {
	// TODO: Intercept the integration's name/ID from gRPC.
	scope := "scope"

	token, err := s.impl.Create(ctx, scope, req.Msg.Data, req.Msg.Key)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&secretsv1.CreateResponse{Token: token}), nil
}

// Get retrieves a connection's key-value data based on the given token.
// If the token isn’t found then we return an error.
func (s server) Get(ctx context.Context, req *connect.Request[secretsv1.GetRequest]) (*connect.Response[secretsv1.GetResponse], error) {
	// TODO: Intercept the integration's name/ID from gRPC.
	scope := "scope"

	data, err := s.impl.Get(ctx, scope, req.Msg.Token)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&secretsv1.GetResponse{Data: data}), nil
}

// List enumerates all the tokens (0 or more) that are associated with a given
// connection identifier. This enables autokitteh to dispatch/fan-out asynchronous
// events that it receives from integrations through all the relevant connections.
func (s server) List(ctx context.Context, req *connect.Request[secretsv1.ListRequest]) (*connect.Response[secretsv1.ListResponse], error) {
	// TODO: Intercept the integration's name/ID from gRPC.
	scope := "scope"

	tokens, err := s.impl.List(ctx, scope, req.Msg.Key)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&secretsv1.ListResponse{Tokens: tokens}), nil
}
