package buildsgrpcsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	buildsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1/buildsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	builds sdkservices.Builds

	buildsv1connect.UnimplementedBuildsServiceHandler
}

var _ buildsv1connect.BuildsServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, builds sdkservices.Builds) {
	srv := server{builds: builds}

	path, namer := buildsv1connect.NewBuildsServiceHandler(&srv)
	muxes.Auth.Handle(path, namer)
}

func (s *server) Get(ctx context.Context, req *connect.Request[buildsv1.GetRequest]) (*connect.Response[buildsv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	buildID, err := sdktypes.StrictParseBuildID(msg.BuildId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	build, err := s.builds.Get(ctx, buildID)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&buildsv1.GetResponse{Build: build.ToProto()}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[buildsv1.ListRequest]) (*connect.Response[buildsv1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	filter := sdkservices.ListBuildsFilter{
		Limit:     msg.Limit,
		ProjectID: pid,
	}

	builds, err := s.builds.List(ctx, filter)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	buildsPB := kittehs.Transform(builds, sdktypes.ToProto)

	return connect.NewResponse(&buildsv1.ListResponse{Builds: buildsPB}), nil
}

func (s *server) Download(ctx context.Context, req *connect.Request[buildsv1.DownloadRequest]) (*connect.Response[buildsv1.DownloadResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	buildID, err := sdktypes.StrictParseBuildID(msg.BuildId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	data, err := s.builds.Download(ctx, buildID)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	defer data.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(data); err != nil {
		return nil, sdkerrors.AsConnectError(fmt.Errorf("server error: %w", err))
	}

	return connect.NewResponse(&buildsv1.DownloadResponse{Data: buf.Bytes()}), nil
}

func (s *server) Save(ctx context.Context, req *connect.Request[buildsv1.SaveRequest]) (*connect.Response[buildsv1.SaveResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	build, err := sdktypes.BuildFromProto(msg.Build)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	bid, err := s.builds.Save(ctx, build, msg.Data)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&buildsv1.SaveResponse{BuildId: bid.String()}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[buildsv1.DeleteRequest]) (*connect.Response[buildsv1.DeleteResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	bid, err := sdktypes.ParseBuildID(msg.BuildId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err = s.builds.Delete(ctx, bid); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&buildsv1.DeleteResponse{}), nil
}

func (s *server) Describe(ctx context.Context, req *connect.Request[buildsv1.DescribeRequest]) (*connect.Response[buildsv1.DescribeResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	bid, err := sdktypes.ParseBuildID(msg.BuildId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	bf, err := s.builds.Describe(ctx, bid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	desc, err := json.Marshal(bf)
	if err != nil {
		return nil, sdkerrors.AsConnectError(fmt.Errorf("marshal: %w", err))
	}

	return connect.NewResponse(&buildsv1.DescribeResponse{DescriptionJson: string(desc)}), nil
}
