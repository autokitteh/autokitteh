package integrationsgrpcsvc

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1/integrationsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	integrations sdkservices.Integrations

	integrationsv1connect.UnimplementedIntegrationsServiceHandler
}

var _ integrationsv1connect.IntegrationsServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, integrations sdkservices.Integrations) {
	s := server{integrations: integrations}
	path, handler := integrationsv1connect.NewIntegrationsServiceHandler(&s)
	mux.Handle(path, handler)
}

func (s *server) Get(ctx context.Context, req *connect.Request[integrationsv1.GetRequest]) (*connect.Response[integrationsv1.GetResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	id, err := sdktypes.StrictParseIntegrationID(req.Msg.IntegrationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("integration_id: %w", err))
	}

	i, err := s.integrations.Get(ctx, id)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&integrationsv1.GetResponse{Integration: i.Get().ToProto()}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[integrationsv1.ListRequest]) (*connect.Response[integrationsv1.ListResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	// TODO: Tags

	is, err := s.integrations.List(ctx, req.Msg.NameSubstring)
	if err != nil {
		return nil, toConnectError(err)
	}

	return connect.NewResponse(&integrationsv1.ListResponse{
		Integrations: kittehs.Transform(is, sdktypes.ToProto),
	}), nil
}

func (*server) Call(context.Context, *connect.Request[integrationsv1.CallRequest]) (*connect.Response[integrationsv1.CallResponse], error) {
	// TODO
	return nil, connect.NewError(connect.CodeUnimplemented, sdkerrors.ErrNotImplemented)
}

func (*server) Configure(context.Context, *connect.Request[integrationsv1.ConfigureRequest]) (*connect.Response[integrationsv1.ConfigureResponse], error) {
	// TODO
	return nil, connect.NewError(connect.CodeUnimplemented, sdkerrors.ErrNotImplemented)
}

func toConnectError(err error) *connect.Error {
	switch {
	case errors.Is(err, sdkerrors.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, sdkerrors.ErrUnauthenticated):
		return connect.NewError(connect.CodeUnauthenticated, err)
	case errors.Is(err, sdkerrors.ErrUnauthorized):
		return connect.NewError(connect.CodePermissionDenied, err)
	default:
		return connect.NewError(connect.CodeUnknown, err)
	}
}
