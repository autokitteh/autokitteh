package authgrpcsvc

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/proto"
	authv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/auth/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/auth/v1/authv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type server struct {
	auth sdkservices.Auth
	authv1connect.UnimplementedAuthServiceHandler
}

var _ authv1connect.AuthServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, auth sdkservices.Auth) {
	srv := server{auth: auth}

	path, namer := authv1connect.NewAuthServiceHandler(&srv)
	mux.Handle(path, namer)
}

func (s *server) WhoAmI(ctx context.Context, req *connect.Request[authv1.WhoAmIRequest]) (*connect.Response[authv1.WhoAmIResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	u, err := s.auth.WhoAmI(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	return connect.NewResponse(&authv1.WhoAmIResponse{User: u.ToProto()}), nil
}
