package sdkintegrationsclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1/integrationsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct {
	desc   sdktypes.Integration
	client integrationsv1connect.IntegrationsServiceClient
}

func (i *integration) Get() sdktypes.Integration { return i.desc }

func (i *integration) Configure(ctx context.Context, cid sdktypes.ConnectionID) (map[string]sdktypes.Value, map[string]string, error) {
	resp, err := i.client.Configure(ctx, connect.NewRequest(&integrationsv1.ConfigureRequest{
		IntegrationId: i.desc.ID().String(),
		ConnectionId:  cid.String(),
	}))
	if err != nil {
		return nil, nil, rpcerrors.ToSDKError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return nil, nil, err
	}

	vs, err := kittehs.TransformMapValuesError(resp.Msg.Values, sdktypes.StrictValueFromProto)
	if err != nil {
		return nil, nil, err
	}

	return vs, resp.Msg.Config, nil
}

func (i *integration) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	intid := i.desc.ID()

	if v.GetFunction().ExecutorID().ToIntegrationID() != intid {
		return sdktypes.InvalidValue, fmt.Errorf("function value %v is not from this integration", v)
	}

	req := connect.NewRequest(&integrationsv1.CallRequest{
		IntegrationId: intid.String(),
		Function:      v.ToProto(),
		Args:          kittehs.Transform(args, sdktypes.ToProto),
		Kwargs:        kittehs.TransformMapValues(kwargs, sdktypes.ToProto),
	})
	resp, err := i.client.Call(ctx, req)
	if err != nil {
		return sdktypes.InvalidValue, rpcerrors.ToSDKError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidValue, err
	}

	retv, err := sdktypes.ValueFromProto(resp.Msg.Value)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	perr, err := sdktypes.ProgramErrorFromProto(resp.Msg.Error)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return retv, perr.ToError()
}

func (i *integration) TestConnection(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
	intid := i.desc.ID()

	resp, err := i.client.TestConnection(ctx, connect.NewRequest(&integrationsv1.TestConnectionRequest{
		IntegrationId: intid.String(),
		ConnectionId:  cid.String(),
	}))
	if err != nil {
		return sdktypes.InvalidStatus, rpcerrors.ToSDKError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidStatus, err
	}

	return sdktypes.StrictStatusFromProto(resp.Msg.Status)
}

func (i *integration) GetConnectionStatus(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
	intid := i.desc.ID()

	resp, err := i.client.GetConnectionStatus(ctx, connect.NewRequest(&integrationsv1.GetConnectionStatusRequest{
		IntegrationId: intid.String(),
		ConnectionId:  cid.String(),
	}))
	if err != nil {
		return sdktypes.InvalidStatus, rpcerrors.ToSDKError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidStatus, err
	}

	return sdktypes.StrictStatusFromProto(resp.Msg.Status)
}

func (i *integration) GetConnectionConfig(ctx context.Context, cid sdktypes.ConnectionID) (map[string]string, error) {
	intid := i.desc.ID()

	resp, err := i.client.GetConnectionConfig(ctx, connect.NewRequest(&integrationsv1.GetConnectionConfigRequest{
		IntegrationId: intid.String(),
		ConnectionId:  cid.String(),
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return resp.Msg.Config, nil
}
