package orgsgrpcsvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

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

	orgsv1connect.UnimplementedOrgsServiceHandler
}

var _ orgsv1connect.OrgsServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, orgs sdkservices.Orgs) {
	srv := server{orgs: orgs}

	path, handler := orgsv1connect.NewOrgsServiceHandler(&srv)
	mux.Handle(path, handler)
}

func (s *server) Create(ctx context.Context, req *connect.Request[orgsv1.CreateRequest]) (*connect.Response[orgsv1.CreateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	org, err := sdktypes.OrgFromProto(msg.Org)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	oid, err := s.orgs.Create(ctx, org)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrUnauthorized) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}

		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&orgsv1.CreateResponse{OrgId: oid.String()}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[orgsv1.GetRequest]) (*connect.Response[orgsv1.GetResponse], error) {
	toResponse := func(org sdktypes.Org, err error) (*connect.Response[orgsv1.GetResponse], error) {
		if err != nil {
			if errors.Is(err, sdkerrors.ErrNotFound) {
				return connect.NewResponse(&orgsv1.GetResponse{}), nil
			} else if errors.Is(err, sdkerrors.ErrUnauthorized) {
				return nil, connect.NewError(connect.CodePermissionDenied, err)
			}

			return nil, connect.NewError(connect.CodeUnknown, err)
		}

		return connect.NewResponse(&orgsv1.GetResponse{Org: org.ToProto()}), nil
	}

	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	oid, err := sdktypes.ParseOrgID(msg.OrgId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("org_id: %w", err))
	}

	if oid.IsValid() {
		return toResponse(s.orgs.GetByID(ctx, oid))
	}

	// a name must've been supplied here.
	n, err := sdktypes.Strict(sdktypes.ParseSymbol(msg.Name))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name: %w", err))
	}

	return toResponse(s.orgs.GetByName(ctx, n))
}

func (s *server) ListMembers(ctx context.Context, req *connect.Request[orgsv1.ListMembersRequest]) (*connect.Response[orgsv1.ListMembersResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	oid, err := sdktypes.StrictParseOrgID(msg.OrgId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("org_id: %w", err))
	}

	us, err := s.orgs.ListMembers(ctx, oid)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		} else if errors.Is(err, sdkerrors.ErrUnauthorized) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}

		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	pbus := kittehs.Transform(us, sdktypes.ToProto)

	return connect.NewResponse(&orgsv1.ListMembersResponse{Users: pbus}), nil
}

func (s *server) AddMember(ctx context.Context, req *connect.Request[orgsv1.AddMemberRequest]) (*connect.Response[orgsv1.AddMemberResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	oid, err := sdktypes.StrictParseOrgID(msg.OrgId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("org_id: %w", err))
	}

	uid, err := sdktypes.StrictParseUserID(msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id: %w", err))
	}

	if err := s.orgs.AddMember(ctx, oid, uid); err != nil {
		if errors.Is(err, sdkerrors.ErrAlreadyExists) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		} else if errors.Is(err, sdkerrors.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		} else if errors.Is(err, sdkerrors.ErrUnauthorized) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}

		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&orgsv1.AddMemberResponse{}), nil
}

func (s *server) RemoveMember(ctx context.Context, req *connect.Request[orgsv1.RemoveMemberRequest]) (*connect.Response[orgsv1.RemoveMemberResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	oid, err := sdktypes.StrictParseOrgID(msg.OrgId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("org_id: %w", err))
	}

	uid, err := sdktypes.StrictParseUserID(msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id: %w", err))
	}

	if err := s.orgs.RemoveMember(ctx, oid, uid); err != nil {
		if errors.Is(err, sdkerrors.ErrUnauthorized) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}

		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&orgsv1.RemoveMemberResponse{}), nil
}

func (s *server) IsMember(ctx context.Context, req *connect.Request[orgsv1.IsMemberRequest]) (*connect.Response[orgsv1.IsMemberResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	oid, err := sdktypes.StrictParseOrgID(msg.OrgId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("org_id: %w", err))
	}

	uid, err := sdktypes.StrictParseUserID(msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id: %w", err))
	}

	isMember, err := s.orgs.IsMember(ctx, oid, uid)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrUnauthorized) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}

		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&orgsv1.IsMemberResponse{IsMember: isMember}), nil
}

func (s *server) ListUserMemberships(ctx context.Context, req *connect.Request[orgsv1.ListUserMembershipsRequest]) (*connect.Response[orgsv1.ListUserMembershipsResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	uid, err := sdktypes.ParseUserID(msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id: %w", err))
	}

	os, err := s.orgs.ListUserMemberships(ctx, uid)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		} else if errors.Is(err, sdkerrors.ErrUnauthorized) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}

		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	pbos := kittehs.Transform(os, sdktypes.ToProto)

	return connect.NewResponse(&orgsv1.ListUserMembershipsResponse{Orgs: pbos}), nil
}
