package sdkorgsclient

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	orgsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1/orgsv1connect"

	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client orgsv1connect.OrgsServiceClient
}

func New(p sdkclient.Params) sdkservices.Orgs {
	return &client{client: internal.New(orgsv1connect.NewOrgsServiceClient, p)}
}

func (c *client) Create(ctx context.Context, o sdktypes.Org) (sdktypes.OrgID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&orgsv1.CreateRequest{
		Org: o.ToProto(),
	}))
	if err != nil {
		return sdktypes.InvalidOrgID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidOrgID, err
	}

	return sdktypes.ParseOrgID(resp.Msg.OrgId)
}

func (c *client) GetByID(ctx context.Context, oid sdktypes.OrgID) (sdktypes.Org, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&orgsv1.GetRequest{
		OrgId: oid.String(),
	}))
	if err != nil {
		return sdktypes.InvalidOrg, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidOrg, err
	}

	return sdktypes.OrgFromProto(resp.Msg.Org)
}

func (c *client) Update(ctx context.Context, u sdktypes.Org) error {
	resp, err := c.client.Update(ctx, connect.NewRequest(&orgsv1.UpdateRequest{
		Org: u.ToProto(),
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}

func (c *client) ListMembers(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.UserID, error) {
	resp, err := c.client.ListMembers(ctx, connect.NewRequest(&orgsv1.ListMembersRequest{
		OrgId: oid.String(),
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.UserIds, sdktypes.StrictParseUserID)
}

func (c *client) AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	resp, err := c.client.AddMember(ctx, connect.NewRequest(&orgsv1.AddMemberRequest{
		OrgId:  oid.String(),
		UserId: uid.String(),
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) RemoveMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	resp, err := c.client.RemoveMember(ctx, connect.NewRequest(&orgsv1.RemoveMemberRequest{
		OrgId:  oid.String(),
		UserId: uid.String(),
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) IsMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (bool, error) {
	resp, err := c.client.IsMember(ctx, connect.NewRequest(&orgsv1.IsMemberRequest{
		OrgId:  oid.String(),
		UserId: uid.String(),
	}))
	if err != nil {
		return false, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return false, err
	}

	return resp.Msg.IsMember, nil
}

func (c *client) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.Org, error) {
	resp, err := c.client.GetOrgsForUser(ctx, connect.NewRequest(&orgsv1.GetOrgsForUserRequest{
		UserId: uid.String(),
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Orgs, sdktypes.OrgFromProto)
}
