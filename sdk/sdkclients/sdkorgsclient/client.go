package sdkorgsclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

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

func (c *client) Delete(ctx context.Context, oid sdktypes.OrgID) error {
	resp, err := c.client.Delete(ctx, connect.NewRequest(&orgsv1.DeleteRequest{
		OrgId: oid.String(),
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) Update(ctx context.Context, u sdktypes.Org, fm *sdktypes.FieldMask) error {
	resp, err := c.client.Update(ctx, connect.NewRequest(&orgsv1.UpdateRequest{
		Org:       u.ToProto(),
		FieldMask: fm,
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}

func (c *client) ListMembers(ctx context.Context, oid sdktypes.OrgID) ([]*sdkservices.UserIDWithMemberStatus, error) {
	resp, err := c.client.ListMembers(ctx, connect.NewRequest(&orgsv1.ListMembersRequest{
		OrgId: oid.String(),
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Members, func(o *orgsv1.OrgMember) (*sdkservices.UserIDWithMemberStatus, error) {
		s, err := sdktypes.OrgMemberStatusFromProto(o.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to parse org member status: %w", err)
		}

		uid, err := sdktypes.StrictParseUserID(o.UserId)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user id: %w", err)
		}

		return &sdkservices.UserIDWithMemberStatus{UserID: uid, Status: s}, nil
	})
}

func (c *client) AddMember(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID, status sdktypes.OrgMemberStatus) error {
	resp, err := c.client.AddMember(ctx, connect.NewRequest(&orgsv1.AddMemberRequest{
		Member: &orgsv1.OrgMember{
			OrgId:  oid.String(),
			UserId: uid.String(),
			Status: status.ToProto(),
		},
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

func (c *client) GetMemberStatus(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID) (sdktypes.OrgMemberStatus, error) {
	resp, err := c.client.GetMember(ctx, connect.NewRequest(&orgsv1.GetMemberRequest{
		OrgId:  oid.String(),
		UserId: uid.String(),
	}))
	if err != nil {
		return sdktypes.OrgMemberStatusUnspecified, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.OrgMemberStatusUnspecified, err
	}

	return sdktypes.OrgMemberStatusFromProto(resp.Msg.Member.Status)
}

func (c *client) GetOrgsForUser(ctx context.Context, uid sdktypes.UserID) ([]*sdkservices.OrgWithMemberStatus, error) {
	resp, err := c.client.GetOrgsForUser(ctx, connect.NewRequest(&orgsv1.GetOrgsForUserRequest{
		UserId:      uid.String(),
		IncludeOrgs: true,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Members, func(o *orgsv1.OrgMember) (*sdkservices.OrgWithMemberStatus, error) {
		s, err := sdktypes.OrgMemberStatusFromProto(o.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to parse org member status: %w", err)
		}

		org, err := sdktypes.OrgFromProto(resp.Msg.Orgs[o.OrgId])
		if err != nil {
			return nil, fmt.Errorf("failed to parse org: %w", err)
		}

		return &sdkservices.OrgWithMemberStatus{Org: org, Status: s}, nil
	})
}

func (c *client) UpdateMemberStatus(ctx context.Context, oid sdktypes.OrgID, uid sdktypes.UserID, status sdktypes.OrgMemberStatus) error {
	resp, err := c.client.UpdateMember(ctx, connect.NewRequest(&orgsv1.UpdateMemberRequest{
		Member: &orgsv1.OrgMember{
			OrgId:  oid.String(),
			UserId: uid.String(),
			Status: status.ToProto(),
		},
		FieldMask: kittehs.Must1(fieldmaskpb.New(&orgsv1.OrgMember{}, "status")),
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}
