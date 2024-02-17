package sdksecretsclient

import (
	"context"

	"connectrpc.com/connect"

	secretsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/secrets/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/secrets/v1/secretsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type client struct {
	client secretsv1connect.SecretsServiceClient
}

func New(p sdkclient.Params) sdkservices.Secrets {
	return client{client: internal.New(secretsv1connect.NewSecretsServiceClient, p)}
}

func (c client) Create(ctx context.Context, scope string, data map[string]string, key string) (string, error) {
	req := &secretsv1.CreateRequest{Data: data, Key: key}
	resp, err := c.client.Create(ctx, connect.NewRequest(req))
	if err != nil {
		return "", err
	}
	return resp.Msg.Token, nil
}

func (c client) Get(ctx context.Context, scope string, token string) (map[string]string, error) {
	req := &secretsv1.GetRequest{Token: token}
	resp, err := c.client.Get(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg.Data, nil
}

func (c client) List(ctx context.Context, scope string, key string) ([]string, error) {
	req := &secretsv1.ListRequest{Key: key}
	resp, err := c.client.List(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg.Tokens, nil
}
