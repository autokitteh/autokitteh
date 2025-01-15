package sdkusersclient

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	usersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1/usersv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client usersv1connect.UsersServiceClient
}

func New(p sdkclient.Params) sdkservices.Users {
	return &client{client: internal.New(usersv1connect.NewUsersServiceClient, p)}
}

func (c *client) Create(ctx context.Context, u sdktypes.User) (sdktypes.UserID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&usersv1.CreateRequest{
		User: u.ToProto(),
	}))
	if err != nil {
		return sdktypes.InvalidUserID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidUserID, err
	}

	return sdktypes.ParseUserID(resp.Msg.UserId)
}

func (c *client) Get(ctx context.Context, uid sdktypes.UserID, email string) (sdktypes.User, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&usersv1.GetRequest{
		UserId: uid.String(),
		Email:  email,
	}))
	if err != nil {
		return sdktypes.InvalidUser, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidUser, err
	}

	return sdktypes.UserFromProto(resp.Msg.User)
}

func (c *client) BatchGetByIDs(ctx context.Context, uids []sdktypes.UserID) ([]sdktypes.User, error) {
	resp, err := c.client.BatchGet(ctx, connect.NewRequest(&usersv1.BatchGetRequest{
		UserIds: kittehs.TransformToStrings(uids),
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Users, sdktypes.UserFromProto)
}

func (c *client) GetID(ctx context.Context, email string) (sdktypes.UserID, error) {
	resp, err := c.client.GetID(ctx, connect.NewRequest(&usersv1.GetIDRequest{
		Email: email,
	}))
	if err != nil {
		return sdktypes.InvalidUserID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidUserID, err
	}

	return sdktypes.ParseUserID(resp.Msg.UserId)
}

func (c *client) Update(ctx context.Context, u sdktypes.User, fm *sdktypes.FieldMask) error {
	resp, err := c.client.Update(ctx, connect.NewRequest(&usersv1.UpdateRequest{
		User:      u.ToProto(),
		FieldMask: fm,
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}
