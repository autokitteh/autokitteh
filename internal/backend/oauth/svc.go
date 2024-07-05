package oauth

import (
	"context"

	"connectrpc.com/connect"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/proto"
	oauthv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/oauth/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/oauth/v1/oauthv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	impl   sdkservices.OAuth
	logger *zap.Logger

	oauthv1connect.UnimplementedOAuthServiceHandler
}

var _ oauthv1connect.OAuthServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, l *zap.Logger, oauth sdkservices.OAuth) {
	srv := server{logger: l, impl: oauth}
	path, handler := oauthv1connect.NewOAuthServiceHandler(&srv)
	muxes.Auth.Handle(path, handler)
}

func (s *server) Register(ctx context.Context, req *connect.Request[oauthv1.RegisterRequest]) (*connect.Response[oauthv1.RegisterResponse], error) {
	// Validate & parse the request.
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	cfg := &oauth2.Config{
		ClientID:     req.Msg.Config.ClientId,
		ClientSecret: req.Msg.Config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:       req.Msg.Config.AuthUrl,
			DeviceAuthURL: req.Msg.Config.DeviceAuthUrl,
			TokenURL:      req.Msg.Config.TokenUrl,
			AuthStyle:     oauth2.AuthStyle(req.Msg.Config.AuthStyle),
		},
		RedirectURL: req.Msg.Config.RedirectUrl,
		Scopes:      req.Msg.Config.Scopes,
	}

	// Register a new OAuth handler for the caller's configuration.
	err := s.impl.Register(ctx, req.Msg.Id, cfg, req.Msg.Config.Options)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}
	return connect.NewResponse(&oauthv1.RegisterResponse{}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[oauthv1.GetRequest]) (*connect.Response[oauthv1.GetResponse], error) {
	// Validate the request.
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	// Return the requested OAuth handler configuration.
	cfg, opts, err := s.impl.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}
	c := &oauthv1.OAuthConfig{
		ClientId:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,

		AuthUrl:       cfg.Endpoint.AuthURL,
		DeviceAuthUrl: cfg.Endpoint.DeviceAuthURL,
		TokenUrl:      cfg.Endpoint.TokenURL,
		RedirectUrl:   cfg.RedirectURL,

		AuthStyle: int32(cfg.Endpoint.AuthStyle),
		Options:   opts,
		Scopes:    cfg.Scopes,
	}
	return connect.NewResponse(&oauthv1.GetResponse{Config: c}), nil
}

func (s *server) StartFlow(ctx context.Context, req *connect.Request[oauthv1.StartFlowRequest]) (*connect.Response[oauthv1.StartFlowResponse], error) {
	// Validate the request.
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	cid, err := sdktypes.ParseConnectionID(req.Msg.ConnectionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Redirect the caller to the URL that starts the OAuth flow.
	url, err := s.impl.StartFlow(ctx, req.Msg.Integration, cid, req.Msg.Origin)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}
	return connect.NewResponse(&oauthv1.StartFlowResponse{Url: url}), nil
}

func (s *server) Exchange(ctx context.Context, req *connect.Request[oauthv1.ExchangeRequest]) (*connect.Response[oauthv1.ExchangeResponse], error) {
	// Validate the request.
	if err := proto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	// Return the exchanged OAuth token, based on the authorization code.
	token, err := s.impl.Exchange(ctx, req.Msg.Integration, req.Msg.Code)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&oauthv1.ExchangeResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry.UnixMicro(),
	}), nil
}
