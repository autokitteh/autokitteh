package sdksessionsclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1/sessionsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client sessionsv1connect.SessionsServiceClient
}

func New(p sdkclient.Params) sdkservices.Sessions {
	return &client{client: internal.New(sessionsv1connect.NewSessionsServiceClient, p)}
}

func (c *client) Start(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
	resp, err := c.client.Start(ctx, connect.NewRequest(&sessionsv1.StartRequest{
		Session: session.ToProto(),
	}))
	if err != nil {
		return sdktypes.InvalidSessionID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidSessionID, err
	}

	pid, err := sdktypes.StrictParseSessionID(resp.Msg.SessionId)
	if err != nil {
		return sdktypes.InvalidSessionID, fmt.Errorf("invalid session id: %w", err)
	}

	return pid, nil
}

func (c *client) Stop(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool) error {
	resp, err := c.client.Stop(ctx, connect.NewRequest(&sessionsv1.StopRequest{
		SessionId: sessionID.String(),
		Reason:    reason,
		Terminate: force,
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) Get(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.Session, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&sessionsv1.GetRequest{
		SessionId: sessionID.String(),
	}))
	if err != nil {
		return sdktypes.InvalidSession, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidSession, err
	}

	return sdktypes.SessionFromProto(resp.Msg.Session)
}

func (c *client) GetLog(ctx context.Context, sessionID sdktypes.SessionID) (sdktypes.SessionLog, error) {
	resp, err := c.client.GetLog(ctx, connect.NewRequest(&sessionsv1.GetLogRequest{
		SessionId: sessionID.String(),
	}))
	if err != nil {
		return sdktypes.InvalidSessionLog, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidSessionLog, err
	}

	return sdktypes.SessionLogFromProto(resp.Msg.Log)
}

func (c *client) List(ctx context.Context, filter sdkservices.ListSessionsFilter) (sdkservices.ListSessionResult, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&sessionsv1.ListRequest{
		DeploymentId: filter.DeploymentID.String(),
		EnvId:        filter.EnvID.String(),
		EventId:      filter.EventID.String(),
		BuildId:      filter.BuildID.String(),
		StateType:    filter.StateType.ToProto(),
		CountOnly:    filter.CountOnly,
		PageSize:     filter.PageSize,
		Skip:         filter.Skip,
		PageToken:    filter.PageToken,
	}))
	if err != nil {
		return sdkservices.ListSessionResult{}, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdkservices.ListSessionResult{}, err
	}

	xs, err := kittehs.TransformError(resp.Msg.Sessions, sdktypes.SessionFromProto)

	res := sdkservices.ListSessionResult{
		Sessions: xs,
		PaginationResult: sdktypes.PaginationResult{
			TotalCount:    resp.Msg.Count,
			NextPageToken: resp.Msg.NextPageToken,
		},
	}
	return res, err
}

func (c *client) Delete(ctx context.Context, sessionID sdktypes.SessionID) error {
	resp, err := c.client.Delete(ctx, connect.NewRequest(&sessionsv1.DeleteRequest{SessionId: sessionID.String()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}
