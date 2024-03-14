package sdkusersclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	usersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1/usersv1connect"

	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client usersv1connect.UsersServiceClient
}

func New(p sdkclient.Params) sdkservices.Users {
	return &client{client: internal.New(usersv1connect.NewUsersServiceClient, p)}
}

func (c *client) Create(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&usersv1.CreateRequest{
		User: user.ToProto(),
	}))
	if err != nil {
		return sdktypes.InvalidUserID, sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidUserID, err
	}

	oid, err := sdktypes.StrictParseUserID(resp.Msg.UserId)
	if err != nil {
		return sdktypes.InvalidUserID, fmt.Errorf("invalid user: %w", err)
	}

	return oid, nil
}

func (c *client) get(ctx context.Context, req *usersv1.GetRequest) (sdktypes.User, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(req))
	if err != nil {
		return sdktypes.InvalidUser, sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidUser, err
	}

	pbuser := resp.Msg.User

	if pbuser == nil {
		return sdktypes.InvalidUser, nil
	}

	user, err := sdktypes.UserFromProto(pbuser)
	if err != nil {
		return sdktypes.InvalidUser, fmt.Errorf("invalid user: %w", err)
	}

	return user, nil
}

func (c *client) GetByID(ctx context.Context, oid sdktypes.UserID) (sdktypes.User, error) {
	return c.get(ctx, &usersv1.GetRequest{UserId: oid.String()})
}

func (c *client) GetByName(ctx context.Context, n sdktypes.Symbol) (sdktypes.User, error) {
	return c.get(ctx, &usersv1.GetRequest{Name: n.String()})
}
