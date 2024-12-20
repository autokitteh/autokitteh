package orgsgrpcsvc

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	orgsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1/orgsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	orgs sdkservices.Orgs
}

var _ orgsv1connect.OrgsServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, orgs sdkservices.Orgs) {
	srv := server{orgs: orgs}

	path, handler := orgsv1connect.NewOrgsServiceHandler(&srv)
	muxes.Auth.Handle(path, handler)
}

func (s *server) Create(ctx context.Context, req *connect.Request[orgsv1.CreateRequest]) (*connect.Response[orgsv1.CreateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	o, err := sdktypes.Strict(sdktypes.OrgFromProto(msg.Org))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	oid, err := s.orgs.Create(ctx, o)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.CreateResponse{OrgId: oid.String()}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[orgsv1.GetRequest]) (*connect.Response[orgsv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	var o sdktypes.Org

	if req.Msg.OrgId != "" {
		oid, err := sdktypes.ParseOrgID(msg.OrgId)
		if err != nil {
			return nil, sdkerrors.AsConnectError(err)
		}

		if o, err = s.orgs.GetByID(ctx, oid); err != nil {
			return nil, sdkerrors.AsConnectError(err)
		}
	} else if req.Msg.Name != "" {
		n, err := sdktypes.ParseSymbol(msg.Name)
		if err != nil {
			return nil, sdkerrors.AsConnectError(err)
		}

		if o, err = s.orgs.GetByName(ctx, n); err != nil {
			return nil, sdkerrors.AsConnectError(err)
		}
	} else {
		return nil, sdkerrors.NewInvalidArgumentError("either org ID or name must be provided")
	}

	return connect.NewResponse(&orgsv1.GetResponse{Org: o.ToProto()}), nil
}

func (s *server) Update(ctx context.Context, req *connect.Request[orgsv1.UpdateRequest]) (*connect.Response[orgsv1.UpdateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	o, err := sdktypes.OrgFromProto(msg.Org)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err = s.orgs.Update(ctx, o); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.UpdateResponse{}), nil
}

func (s *server) AddMember(ctx context.Context, req *connect.Request[orgsv1.AddMemberRequest]) (*connect.Response[orgsv1.AddMemberResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	oid, err := sdktypes.Strict(sdktypes.ParseOrgID(msg.OrgId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := sdktypes.Strict(sdktypes.ParseUserID(msg.UserId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.orgs.AddMember(ctx, oid, uid); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.AddMemberResponse{}), nil
}

func (s *server) RemoveMember(ctx context.Context, req *connect.Request[orgsv1.RemoveMemberRequest]) (*connect.Response[orgsv1.RemoveMemberResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	oid, err := sdktypes.Strict(sdktypes.ParseOrgID(msg.OrgId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := sdktypes.Strict(sdktypes.ParseUserID(msg.UserId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.orgs.RemoveMember(ctx, oid, uid); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.RemoveMemberResponse{}), nil
}

func (s *server) ListMembers(ctx context.Context, req *connect.Request[orgsv1.ListMembersRequest]) (*connect.Response[orgsv1.ListMembersResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	oid, err := sdktypes.ParseOrgID(msg.OrgId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uids, err := s.orgs.ListMembers(ctx, oid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.ListMembersResponse{
		UserIds: kittehs.TransformToStrings(uids),
	}), nil
}

func (s *server) GetOrgsForUser(ctx context.Context, req *connect.Request[orgsv1.GetOrgsForUserRequest]) (*connect.Response[orgsv1.GetOrgsForUserResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := sdktypes.ParseUserID(msg.UserId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	oids, err := s.orgs.GetOrgsForUser(ctx, uid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.GetOrgsForUserResponse{
		OrgIds: kittehs.TransformToStrings(oids),
	}), nil
}

func (s *server) IsMember(ctx context.Context, req *connect.Request[orgsv1.IsMemberRequest]) (*connect.Response[orgsv1.IsMemberResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := sdktypes.ParseUserID(msg.UserId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	oid, err := sdktypes.ParseOrgID(msg.OrgId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	member, err := s.orgs.IsMember(ctx, oid, uid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.IsMemberResponse{IsMember: member}), nil
}
