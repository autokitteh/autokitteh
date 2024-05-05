package deploymentsgrpcsvc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1/deploymentsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	deployments sdkservices.Deployments

	deploymentsv1connect.UnimplementedDeploymentsServiceHandler
}

var _ deploymentsv1connect.DeploymentsServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, deployments sdkservices.Deployments) {
	srv := server{deployments: deployments}

	path, namer := deploymentsv1connect.NewDeploymentsServiceHandler(&srv)
	muxes.API.Handle(path, namer)
}

func (s *server) Create(ctx context.Context, req *connect.Request[deploymentsv1.CreateRequest]) (*connect.Response[deploymentsv1.CreateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	deployment, err := sdktypes.DeploymentFromProto(msg.Deployment)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	did, err := s.deployments.Create(ctx, deployment)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	return connect.NewResponse(&deploymentsv1.CreateResponse{DeploymentId: did.String()}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[deploymentsv1.ListRequest]) (*connect.Response[deploymentsv1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	filter := sdkservices.ListDeploymentsFilter{Limit: req.Msg.Limit}

	bid, err := sdktypes.ParseBuildID(msg.BuildId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	if bid.IsValid() {
		filter.BuildID = bid
	}

	eid, err := sdktypes.ParseEnvID(msg.EnvId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	if eid.IsValid() {
		filter.EnvID = eid
	}

	state, err := sdktypes.DeploymentStateFromProto(msg.State)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if state != sdktypes.DeploymentStateUnspecified {
		filter.State = state
	}

	filter.IncludeSessionStats = msg.IncludeSessionStats

	deployments, err := s.deployments.List(ctx, filter)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	deploymentsPB := kittehs.Transform(deployments, sdktypes.ToProto)

	return connect.NewResponse(&deploymentsv1.ListResponse{Deployments: deploymentsPB}), nil
}

func (s *server) Activate(ctx context.Context, req *connect.Request[deploymentsv1.ActivateRequest]) (*connect.Response[deploymentsv1.ActivateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	did, err := sdktypes.Strict(sdktypes.ParseDeploymentID(msg.DeploymentId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	err = s.deployments.Activate(ctx, did)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&deploymentsv1.ActivateResponse{}), nil
}

func (s *server) Test(ctx context.Context, req *connect.Request[deploymentsv1.TestRequest]) (*connect.Response[deploymentsv1.TestResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	did, err := sdktypes.Strict(sdktypes.ParseDeploymentID(msg.DeploymentId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	err = s.deployments.Test(ctx, did)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&deploymentsv1.TestResponse{}), nil
}

func (s *server) Drain(ctx context.Context, req *connect.Request[deploymentsv1.DrainRequest]) (*connect.Response[deploymentsv1.DrainResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	did, err := sdktypes.Strict(sdktypes.ParseDeploymentID(msg.DeploymentId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	err = s.deployments.Drain(ctx, did)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&deploymentsv1.DrainResponse{}), nil
}

func (s *server) Deactivate(ctx context.Context, req *connect.Request[deploymentsv1.DeactivateRequest]) (*connect.Response[deploymentsv1.DeactivateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	did, err := sdktypes.Strict(sdktypes.ParseDeploymentID(msg.DeploymentId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	err = s.deployments.Deactivate(ctx, did)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&deploymentsv1.DeactivateResponse{}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[deploymentsv1.GetRequest]) (*connect.Response[deploymentsv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	did, err := sdktypes.Strict(sdktypes.ParseDeploymentID(msg.DeploymentId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	deployment, err := s.deployments.Get(ctx, did)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&deploymentsv1.GetResponse{Deployment: deployment.ToProto()}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[deploymentsv1.DeleteRequest]) (*connect.Response[deploymentsv1.DeleteResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	did, err := sdktypes.Strict(sdktypes.ParseDeploymentID(msg.DeploymentId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err = s.deployments.Delete(ctx, did); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&deploymentsv1.DeleteResponse{}), nil
}
