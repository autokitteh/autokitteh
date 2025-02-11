package sdksessionsclient

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/durationpb"

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

func (c *client) Stop(ctx context.Context, sessionID sdktypes.SessionID, reason string, force bool, forceTimeout time.Duration) error {
	resp, err := c.client.Stop(ctx, connect.NewRequest(&sessionsv1.StopRequest{
		SessionId:        sessionID.String(),
		Reason:           reason,
		Terminate:        force,
		TerminationDelay: durationpb.New(forceTimeout),
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

func (c *client) GetPrints(ctx context.Context, sid sdktypes.SessionID, pagination sdktypes.PaginationRequest) (*sdkservices.GetPrintsResults, error) {
	resp, err := c.client.GetPrints(ctx, connect.NewRequest(&sessionsv1.GetPrintsRequest{
		SessionId: sid.String(),
		PageSize:  pagination.PageSize,
		Skip:      pagination.Skip,
		PageToken: pagination.PageToken,
		Ascending: pagination.Ascending,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	rs, err := kittehs.TransformError(resp.Msg.Prints, func(p *sessionsv1.GetPrintsResponse_Print) (*sdkservices.SessionPrint, error) {
		v, err := sdktypes.ValueFromProto(p.V)
		if err != nil {
			return nil, err
		}

		return &sdkservices.SessionPrint{
			Timestamp: p.T.AsTime(),
			Value:     v,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return &sdkservices.GetPrintsResults{
		Prints: rs,
		PaginationResult: sdktypes.PaginationResult{
			NextPageToken: resp.Msg.NextPageToken,
		},
	}, nil
}

func (c *client) GetLog(ctx context.Context, filter sdkservices.SessionLogRecordsFilter) (*sdkservices.GetLogResults, error) {
	resp, err := c.client.GetLog(ctx, connect.NewRequest(&sessionsv1.GetLogRequest{
		SessionId: filter.SessionID.String(),
		PageSize:  filter.PageSize,
		Skip:      filter.Skip,
		PageToken: filter.PageToken,
		Ascending: filter.Ascending,
		Types:     filter.Types,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	rs, err := kittehs.TransformError(resp.Msg.Records, sdktypes.SessionLogRecordFromProto)
	if err != nil {
		return nil, err
	}

	return &sdkservices.GetLogResults{
		Records: rs,
		PaginationResult: sdktypes.PaginationResult{
			TotalCount:    resp.Msg.Count,
			NextPageToken: resp.Msg.NextPageToken,
		},
	}, nil
}

func (c *client) List(ctx context.Context, filter sdkservices.ListSessionsFilter) (*sdkservices.ListSessionResult, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&sessionsv1.ListRequest{
		DeploymentId: filter.DeploymentID.String(),
		OrgId:        filter.OrgID.String(),
		ProjectId:    filter.ProjectID.String(),
		EventId:      filter.EventID.String(),
		BuildId:      filter.BuildID.String(),
		StateType:    filter.StateType.ToProto(),
		CountOnly:    filter.CountOnly,
		PageSize:     filter.PageSize,
		Skip:         filter.Skip,
		PageToken:    filter.PageToken,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	xs, err := kittehs.TransformError(resp.Msg.Sessions, sdktypes.SessionFromProto)
	if err != nil {
		return nil, err
	}

	return &sdkservices.ListSessionResult{
		Sessions: xs,
		PaginationResult: sdktypes.PaginationResult{
			TotalCount:    resp.Msg.Count,
			NextPageToken: resp.Msg.NextPageToken,
		},
	}, nil
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
