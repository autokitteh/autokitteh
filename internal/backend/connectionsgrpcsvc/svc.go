package connectionsgrpcsvc

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	connectionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1/connectionsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	connections sdkservices.Connections
	connectionsv1connect.UnimplementedConnectionsServiceHandler
}

var _ connectionsv1connect.ConnectionsServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, connections sdkservices.Connections) {
	s := server{connections: connections}
	path, handler := connectionsv1connect.NewConnectionsServiceHandler(&s)
	mux.Handle(path, handler)
}

func (s *server) Create(ctx context.Context, req *connect.Request[connectionsv1.CreateRequest]) (*connect.Response[connectionsv1.CreateResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	c, err := sdktypes.StrictConnectionFromProto(req.Msg.Connection)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	id, err := s.connections.Create(ctx, c)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	return connect.NewResponse(&connectionsv1.CreateResponse{ConnectionId: id.String()}), nil
}

func (s *server) Update(ctx context.Context, req *connect.Request[connectionsv1.UpdateRequest]) (*connect.Response[connectionsv1.UpdateResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	c, err := sdktypes.ConnectionFromProto(req.Msg.Connection)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.connections.Update(ctx, c); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	return connect.NewResponse(&connectionsv1.UpdateResponse{}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[connectionsv1.DeleteRequest]) (*connect.Response[connectionsv1.DeleteResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	id, err := sdktypes.StrictParseConnectionID(req.Msg.ConnectionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	err = s.connections.Delete(ctx, id)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	return connect.NewResponse(&connectionsv1.DeleteResponse{}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[connectionsv1.GetRequest]) (*connect.Response[connectionsv1.GetResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	id, err := sdktypes.StrictParseConnectionID(req.Msg.ConnectionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	c, err := s.connections.Get(ctx, id)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if !c.IsValid() {
		return connect.NewResponse(&connectionsv1.GetResponse{}), nil
	}

	return connect.NewResponse(&connectionsv1.GetResponse{Connection: c.ToProto()}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[connectionsv1.ListRequest]) (*connect.Response[connectionsv1.ListResponse], error) {
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	f := sdkservices.ListConnectionsFilter{
		IntegrationToken: req.Msg.IntegrationToken, // Optional
	}

	iid, err := sdktypes.ParseIntegrationID(req.Msg.IntegrationId) // Optional
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	f.IntegrationID = iid

	pid, err := sdktypes.ParseProjectID(req.Msg.ProjectId) // Optional
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	f.ProjectID = pid

	cs, err := s.connections.List(ctx, f)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	return connect.NewResponse(&connectionsv1.ListResponse{
		Connections: kittehs.Transform(cs, sdktypes.ToProto),
	}), nil
}
