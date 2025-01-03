package usersgrpcsvc

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/proto"
	usersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1/usersv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	users sdkservices.Users

	usersv1connect.UnimplementedUsersServiceHandler
}

var _ usersv1connect.UsersServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, users sdkservices.Users) {
	srv := server{users: users}

	path, handler := usersv1connect.NewUsersServiceHandler(&srv)
	muxes.Auth.Handle(path, handler)
}

func (s *server) Create(ctx context.Context, req *connect.Request[usersv1.CreateRequest]) (*connect.Response[usersv1.CreateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	u, err := sdktypes.Strict(sdktypes.UserFromProto(msg.User))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if msg.User.Disabled {
		if u.Status() != sdktypes.UserStatusUnspecified {
			return nil, sdkerrors.NewInvalidArgumentError("status and disabled are mutually exclusive")
		}

		u = u.WithStatus(sdktypes.UserStatusDisabled)
	}

	uid, err := s.users.Create(ctx, u)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&usersv1.CreateResponse{UserId: uid.String()}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[usersv1.GetRequest]) (*connect.Response[usersv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := sdktypes.ParseUserID(msg.UserId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	u, err := s.users.Get(ctx, uid, msg.Email)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pb := u.ToProto()
	if u.Status() == sdktypes.UserStatusDisabled {
		pb.Disabled = true
	}

	return connect.NewResponse(&usersv1.GetResponse{User: pb}), nil
}

func (s *server) Update(ctx context.Context, req *connect.Request[usersv1.UpdateRequest]) (*connect.Response[usersv1.UpdateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	u, err := sdktypes.UserFromProto(msg.User)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err = s.users.Update(ctx, u, msg.FieldMask); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&usersv1.UpdateResponse{}), nil
}
