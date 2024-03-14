package sdkorgsclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	orgsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1/orgsv1connect"

	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client orgsv1connect.OrgsServiceClient
}

func New(p sdkclient.Params) sdkservices.Orgs {
	return &client{client: internal.New(orgsv1connect.NewOrgsServiceClient, p)}
}

func (c *client) Create(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&orgsv1.CreateRequest{
		Org: org.ToProto(),
	}))
	if err != nil {
		return sdktypes.InvalidOrgID, sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidOrgID, err
	}

	oid, err := sdktypes.StrictParseOrgID(resp.Msg.OrgId)
	if err != nil {
		return sdktypes.InvalidOrgID, fmt.Errorf("invalid org: %w", err)
	}

	return oid, nil
}

func (c *client) get(ctx context.Context, req *orgsv1.GetRequest) (sdktypes.Org, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(req))
	if err != nil {
		return sdktypes.InvalidOrg, sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidOrg, err
	}

	pborg := resp.Msg.Org

	if pborg == nil {
		return sdktypes.InvalidOrg, nil
	}

	org, err := sdktypes.StrictOrgFromProto(pborg)
	if err != nil {
		return sdktypes.InvalidOrg, fmt.Errorf("invalid org: %w", err)
	}

	return org, nil
}

func (c *client) GetByID(ctx context.Context, oid sdktypes.OrgID) (sdktypes.Org, error) {
	return c.get(ctx, &orgsv1.GetRequest{OrgId: oid.String()})
}

func (c *client) GetByName(ctx context.Context, n sdktypes.Symbol) (sdktypes.Org, error) {
	return c.get(ctx, &orgsv1.GetRequest{Name: n.String()})
}

func (c *client) AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) error {
	resp, err := c.client.AddMember(ctx, connect.NewRequest(&orgsv1.AddMemberRequest{
		OrgId:  oid.String(),
		UserId: uid.String(),
	}))
	if err != nil {
		return sdkerrors.AsConnectError(err)
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
		return sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) ListMembers(ctx context.Context, oid sdktypes.OrgID) ([]sdktypes.User, error) {
	resp, err := c.client.ListMembers(ctx, connect.NewRequest(&orgsv1.ListMembersRequest{
		OrgId: oid.String(),
	}))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Users, sdktypes.StrictUserFromProto)
}

func (c *client) IsMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (bool, error) {
	resp, err := c.client.IsMember(ctx, connect.NewRequest(&orgsv1.IsMemberRequest{
		OrgId:  oid.String(),
		UserId: uid.String(),
	}))
	if err != nil {
		return false, sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return false, err
	}

	return resp.Msg.IsMember, nil
}

func (c *client) ListUserMemberships(ctx context.Context, uid sdktypes.UserID) ([]sdktypes.Org, error) {
	resp, err := c.client.ListUserMemberships(
		ctx,
		connect.NewRequest(
			&orgsv1.ListUserMembershipsRequest{UserId: uid.String()},
		),
	)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Orgs, sdktypes.StrictOrgFromProto)
}
