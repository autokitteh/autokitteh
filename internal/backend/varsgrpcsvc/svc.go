package varsgrpcsvc

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	varsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/vars/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/vars/v1/varsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	vars sdkservices.Vars

	varsv1connect.UnimplementedVarsServiceHandler
}

var _ varsv1connect.VarsServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, vars sdkservices.Vars) {
	srv := server{vars: vars}

	path, handler := varsv1connect.NewVarsServiceHandler(&srv)
	muxes.Main.Auth.Handle(path, handler)
}

func (s *server) Set(ctx context.Context, req *connect.Request[varsv1.SetRequest]) (*connect.Response[varsv1.SetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	vs, err := kittehs.TransformError(req.Msg.Vars, sdktypes.StrictVarFromProto)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.vars.Set(ctx, vs...); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&varsv1.SetResponse{}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[varsv1.GetRequest]) (*connect.Response[varsv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sid, err := sdktypes.Strict(sdktypes.ParseVarScopeID(msg.ScopeId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	ns, err := kittehs.TransformError(msg.Names, sdktypes.StrictParseSymbol)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	vs, err := s.vars.Get(ctx, sid, ns...)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&varsv1.GetResponse{
		Vars: kittehs.Transform(vs, sdktypes.ToProto),
	}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[varsv1.DeleteRequest]) (*connect.Response[varsv1.DeleteResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sid, err := sdktypes.Strict(sdktypes.ParseVarScopeID(msg.ScopeId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	ns, err := kittehs.TransformError(msg.Names, sdktypes.StrictParseSymbol)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.vars.Delete(ctx, sid, ns...); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&varsv1.DeleteResponse{}), nil
}

func (s *server) FindConnectionID(ctx context.Context, req *connect.Request[varsv1.FindConnectionIDsRequest]) (*connect.Response[varsv1.FindConnectionIDsResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	iid, err := sdktypes.Strict(sdktypes.ParseIntegrationID(msg.IntegrationId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	n, err := sdktypes.Strict(sdktypes.ParseSymbol(msg.Name))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	cids, err := s.vars.FindConnectionIDs(ctx, iid, n, msg.Value)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&varsv1.FindConnectionIDsResponse{
		ConnectionIds: kittehs.TransformToStrings(cids),
	}), nil
}
