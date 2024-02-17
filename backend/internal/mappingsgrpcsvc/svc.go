package mappingsgrpcsvc

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	mappingsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/mappings/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/mappings/v1/mappingsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	mappings sdkservices.Mappings

	mappingsv1connect.UnimplementedMappingsServiceHandler
}

var _ mappingsv1connect.MappingsServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, mappings sdkservices.Mappings) {
	srv := server{mappings: mappings}

	path, namer := mappingsv1connect.NewMappingsServiceHandler(&srv)
	mux.Handle(path, namer)
}

func (s *server) Create(ctx context.Context, req *connect.Request[mappingsv1.CreateRequest]) (*connect.Response[mappingsv1.CreateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	mapping, err := sdktypes.StrictMappingFromProto(msg.Mapping)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	mid, err := s.mappings.Create(ctx, mapping)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	return connect.NewResponse(&mappingsv1.CreateResponse{MappingId: mid.String()}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[mappingsv1.DeleteRequest]) (*connect.Response[mappingsv1.DeleteResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	mid, err := sdktypes.ParseMappingID(msg.MappingId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := s.mappings.Delete(ctx, mid); err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	return connect.NewResponse(&mappingsv1.DeleteResponse{}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[mappingsv1.GetRequest]) (*connect.Response[mappingsv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	mid, err := sdktypes.ParseMappingID(msg.MappingId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	mapping, err := s.mappings.Get(ctx, mid)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	return connect.NewResponse(&mappingsv1.GetResponse{Mapping: mapping.ToProto()}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[mappingsv1.ListRequest]) (*connect.Response[mappingsv1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	eid, err := sdktypes.ParseEnvID(msg.EnvId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	cid, err := sdktypes.ParseConnectionID(msg.ConnectionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	filter := sdkservices.ListMappingsFilter{
		EnvID:        eid,
		ConnectionID: cid,
	}

	mappings, err := s.mappings.List(ctx, filter)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	mappingsPB := kittehs.Transform(mappings, sdktypes.ToProto)
	return connect.NewResponse(&mappingsv1.ListResponse{Mappings: mappingsPB}), nil
}
