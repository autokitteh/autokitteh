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
	orgs  sdkservices.Orgs
	users sdkservices.Users
}

var _ orgsv1connect.OrgsServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, orgs sdkservices.Orgs, users sdkservices.Users) {
	srv := server{orgs: orgs, users: users}

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

	if req.Msg.OrgId == "" && req.Msg.Name == "" {
		return nil, sdkerrors.NewInvalidArgumentError("either org id or name must be provided")
	}

	oid, err := sdktypes.ParseOrgID(msg.OrgId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	n, err := sdktypes.ParseSymbol(msg.Name)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	var o sdktypes.Org

	switch {
	case oid.IsValid() && n.IsValid():
		err = sdkerrors.NewInvalidArgumentError("only one of org id or name must be provided")
	case oid.IsValid():
		o, err = s.orgs.GetByID(ctx, oid)
	default:
		o, err = s.orgs.GetByName(ctx, n)
	}

	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
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

	if err = s.orgs.Update(ctx, o, msg.FieldMask); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.UpdateResponse{}), nil
}

func (s *server) AddMember(ctx context.Context, req *connect.Request[orgsv1.AddMemberRequest]) (*connect.Response[orgsv1.AddMemberResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	m, err := sdktypes.Strict(sdktypes.OrgMemberFromProto(msg.Member))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.orgs.AddMember(ctx, m); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.AddMemberResponse{}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[orgsv1.DeleteRequest]) (*connect.Response[orgsv1.DeleteResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	oid, err := sdktypes.Strict(sdktypes.ParseOrgID(msg.OrgId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.orgs.Delete(ctx, oid); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.DeleteResponse{}), nil
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

	ms, us, err := s.orgs.ListMembers(ctx, oid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.ListMembersResponse{
		Members: kittehs.Transform(ms, sdktypes.ToProto),
		Users:   kittehs.Transform(us, sdktypes.ToProto),
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

	ms, os, err := s.orgs.GetOrgsForUser(ctx, uid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.GetOrgsForUserResponse{
		Members: kittehs.Transform(ms, sdktypes.ToProto),
		Orgs:    kittehs.Transform(os, sdktypes.ToProto),
	}), nil
}

func (s *server) GetMember(ctx context.Context, req *connect.Request[orgsv1.GetMemberRequest]) (*connect.Response[orgsv1.GetMemberResponse], error) {
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

	m, err := s.orgs.GetMember(ctx, oid, uid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.GetMemberResponse{Member: m.ToProto()}), nil
}

func (s *server) UpdateMember(ctx context.Context, req *connect.Request[orgsv1.UpdateMemberRequest]) (*connect.Response[orgsv1.UpdateMemberResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	member := msg.Member

	m, err := sdktypes.Strict(sdktypes.OrgMemberFromProto(member))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.orgs.UpdateMember(ctx, m, msg.FieldMask); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&orgsv1.UpdateMemberResponse{}), nil
}
