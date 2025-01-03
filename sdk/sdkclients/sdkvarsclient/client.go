package sdkvarsclient

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	varsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/vars/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/vars/v1/varsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client varsv1connect.VarsServiceClient
}

func New(p sdkclient.Params) sdkservices.Vars {
	return &client{client: internal.New(varsv1connect.NewVarsServiceClient, p)}
}

func (c *client) Set(ctx context.Context, vs ...sdktypes.Var) error {
	resp, err := c.client.Set(ctx, connect.NewRequest(&varsv1.SetRequest{
		Vars: kittehs.Transform(vs, sdktypes.ToProto),
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}

func (c *client) Delete(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) error {
	resp, err := c.client.Delete(ctx, connect.NewRequest(&varsv1.DeleteRequest{
		ScopeId: sid.String(),
		Names:   kittehs.TransformToStrings(names),
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}

func (c *client) Get(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&varsv1.GetRequest{
		ScopeId: sid.String(),
		Names:   kittehs.TransformToStrings(names),
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError[*varsv1.Var, sdktypes.Var](resp.Msg.Vars, sdktypes.FromProto)
}

func (c *client) FindConnectionIDs(ctx context.Context, iid sdktypes.IntegrationID, name sdktypes.Symbol, value string) ([]sdktypes.ConnectionID, error) {
	resp, err := c.client.FindConnectionIDs(ctx, connect.NewRequest(&varsv1.FindConnectionIDsRequest{
		IntegrationId: iid.String(),
		Name:          name.String(),
		Value:         value,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.ConnectionIds, sdktypes.StrictParseConnectionID)
}
