package sdkoauthclient

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/oauth2"

	oauthv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/oauth/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/oauth/v1/oauthv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type client struct {
	client oauthv1connect.OAuthServiceClient
}

func New(p sdkclient.Params) sdkservices.OAuth {
	return &client{client: internal.New(oauthv1connect.NewOAuthServiceClient, p)}
}

func (c *client) Register(ctx context.Context, id string, cfg *oauth2.Config, opts map[string]string) error {
	config := &oauthv1.OAuthConfig{
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
	req := &oauthv1.RegisterRequest{Id: id, Config: config}
	_, err := c.client.Register(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}
	return nil
}

func (c *client) Get(ctx context.Context, id string) (*oauth2.Config, map[string]string, error) {
	req := &oauthv1.GetRequest{Id: id}
	resp, err := c.client.Get(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, nil, err
	}
	cfg := &oauth2.Config{
		ClientID:     resp.Msg.Config.ClientId,
		ClientSecret: resp.Msg.Config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:       resp.Msg.Config.AuthUrl,
			DeviceAuthURL: resp.Msg.Config.DeviceAuthUrl,
			TokenURL:      resp.Msg.Config.TokenUrl,
			AuthStyle:     oauth2.AuthStyle(resp.Msg.Config.AuthStyle),
		},
		RedirectURL: resp.Msg.Config.RedirectUrl,
		Scopes:      resp.Msg.Config.Scopes,
	}
	return cfg, resp.Msg.Config.Options, nil
}

func (c *client) StartFlow(ctx context.Context, id string) (string, error) {
	req := &oauthv1.StartFlowRequest{Id: id}
	resp, err := c.client.StartFlow(ctx, connect.NewRequest(req))
	if err != nil {
		return "", err
	}
	return resp.Msg.Url, nil
}

func (c *client) Exchange(ctx context.Context, id, state, code string) (*oauth2.Token, error) {
	req := &oauthv1.ExchangeRequest{Id: id, State: state, Code: code}
	resp, err := c.client.Exchange(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return &oauth2.Token{
		AccessToken:  resp.Msg.AccessToken,
		RefreshToken: resp.Msg.RefreshToken,
		TokenType:    resp.Msg.TokenType,
		Expiry:       time.UnixMicro(resp.Msg.Expiry),
	}, nil
}
