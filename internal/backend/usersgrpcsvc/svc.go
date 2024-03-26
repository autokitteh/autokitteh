package usersgrpcsvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

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

func Init(mux *http.ServeMux, users sdkservices.Users) {
	srv := server{users: users}

	path, handler := usersv1connect.NewUsersServiceHandler(&srv)
	mux.Handle(path, handler)
}

func (s *server) Create(ctx context.Context, req *connect.Request[usersv1.CreateRequest]) (*connect.Response[usersv1.CreateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	user, err := sdktypes.UserFromProto(msg.User)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	uid, err := s.users.Create(ctx, user)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrAlreadyExists) {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		} else if errors.Is(err, sdkerrors.ErrUnauthorized) {
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		}

		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&usersv1.CreateResponse{UserId: uid.String()}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[usersv1.GetRequest]) (*connect.Response[usersv1.GetResponse], error) {
	toResponse := func(user sdktypes.User, err error) (*connect.Response[usersv1.GetResponse], error) {
		if err != nil {
			if errors.Is(err, sdkerrors.ErrNotFound) {
				return connect.NewResponse(&usersv1.GetResponse{}), nil
			} else if errors.Is(err, sdkerrors.ErrUnauthorized) {
				return nil, connect.NewError(connect.CodePermissionDenied, err)
			}

			return nil, connect.NewError(connect.CodeUnknown, err)
		}

		return connect.NewResponse(&usersv1.GetResponse{User: user.ToProto()}), nil
	}

	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	uid, err := sdktypes.ParseUserID(msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id: %w", err))
	}

	if uid.IsValid() {
		return toResponse(s.users.GetByID(ctx, uid))
	}

	h, err := sdktypes.Strict(sdktypes.ParseSymbol(msg.Name))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("handle: %w", err))
	}

	return toResponse(s.users.GetByName(ctx, h))
}
