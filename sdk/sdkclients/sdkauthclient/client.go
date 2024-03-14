package sdkauthclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	authv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/auth/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/auth/v1/authv1connect"

	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client authv1connect.AuthServiceClient
}

func (c *client) WhoAmI(ctx context.Context) (sdktypes.User, error) {
	resp, err := c.client.WhoAmI(ctx, connect.NewRequest(&authv1.WhoAmIRequest{}))
	if err != nil {
		return sdktypes.InvalidUser, sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidUser, err
	}
	if resp.Msg.User == nil {
		return sdktypes.InvalidUser, nil
	}

	user, err := sdktypes.StrictUserFromProto(resp.Msg.User)
	if err != nil {
		return sdktypes.InvalidUser, fmt.Errorf("invalid user: %w", err)
	}
	return user, nil
}

func (c *client) CreateToken(ctx context.Context) (string, error) {
	resp, err := c.client.CreateToken(ctx, connect.NewRequest(&authv1.CreateTokenRequest{}))
	if err != nil {
		return "", sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return "", err
	}

	return resp.Msg.Token, nil
}

func New(p sdkclient.Params) sdkservices.Auth {
	return &client{client: internal.New(authv1connect.NewAuthServiceClient, p)}
}
