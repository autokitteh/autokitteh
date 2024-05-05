package integrationsgrpcsvc

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
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

func Init(muxes *muxes.Muxes, integrations sdkservices.Integrations) {
	s := server{integrations: integrations}
	path, handler := integrationsv1connect.NewIntegrationsServiceHandler(&s)
	muxes.API.Handle(path, handler)
}

func (s *server) Get(ctx context.Context, req *connect.Request[integrationsv1.GetRequest]) (*connect.Response[integrationsv1.GetResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	id, err := sdktypes.StrictParseIntegrationID(req.Msg.IntegrationId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	i, err := s.integrations.Get(ctx, id)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
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
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&integrationsv1.ListResponse{
		Integrations: kittehs.Transform(is, sdktypes.ToProto),
	}), nil
}

func (*server) Call(context.Context, *connect.Request[integrationsv1.CallRequest]) (*connect.Response[integrationsv1.CallResponse], error) {
	// TODO
	return nil, sdkerrors.AsConnectError(sdkerrors.ErrNotImplemented)
}

func (*server) Configure(context.Context, *connect.Request[integrationsv1.ConfigureRequest]) (*connect.Response[integrationsv1.ConfigureResponse], error) {
	// TODO
	return nil, sdkerrors.AsConnectError(sdkerrors.ErrNotImplemented)
}
