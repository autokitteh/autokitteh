package sdkenvsclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	envsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1/envsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client envsv1connect.EnvsServiceClient
}

func New(p sdkclient.Params) sdkservices.Envs {
	return &client{client: internal.New(envsv1connect.NewEnvsServiceClient, p)}
}

func (c *client) List(ctx context.Context, pid sdktypes.ProjectID) ([]sdktypes.Env, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&envsv1.ListRequest{ProjectId: pid.String()}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Envs, sdktypes.StrictEnvFromProto)
}

func (c *client) Create(ctx context.Context, env sdktypes.Env) (sdktypes.EnvID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&envsv1.CreateRequest{Env: env.ToProto()}))
	if err != nil {
		return sdktypes.InvalidEnvID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidEnvID, err
	}

	eid, err := sdktypes.StrictParseEnvID(resp.Msg.EnvId)
	if err != nil {
		return sdktypes.InvalidEnvID, fmt.Errorf("invalid env id: %w", err)
	}

	return eid, nil
}

func (c *client) GetByID(ctx context.Context, eid sdktypes.EnvID) (sdktypes.Env, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(
		&envsv1.GetRequest{EnvId: eid.String()},
	))
	if err != nil {
		return sdktypes.InvalidEnv, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidEnv, err
	}

	env, err := sdktypes.StrictEnvFromProto(resp.Msg.Env)
	if err != nil {
		return sdktypes.InvalidEnv, fmt.Errorf("invalid env: %w", err)
	}

	return env, nil
}

func (c *client) GetByName(ctx context.Context, pid sdktypes.ProjectID, en sdktypes.Symbol) (sdktypes.Env, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(
		&envsv1.GetRequest{
			Name:      en.String(),
			ProjectId: pid.String(),
		},
	))
	if err != nil {
		return sdktypes.InvalidEnv, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidEnv, err
	}

	if resp.Msg.Env == nil {
		return sdktypes.InvalidEnv, nil
	}

	env, err := sdktypes.StrictEnvFromProto(resp.Msg.Env)
	if err != nil {
		return sdktypes.InvalidEnv, fmt.Errorf("invalid env: %w", err)
	}

	return env, nil
}

func (c *client) Remove(ctx context.Context, eid sdktypes.EnvID) error {
	resp, err := c.client.Remove(ctx, connect.NewRequest(
		&envsv1.RemoveRequest{EnvId: eid.String()},
	))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}

func (c *client) Update(ctx context.Context, env sdktypes.Env) error {
	resp, err := c.client.Update(ctx, connect.NewRequest(&envsv1.UpdateRequest{Env: env.ToProto()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}

func (c *client) SetVar(ctx context.Context, ev sdktypes.EnvVar) error {
	resp, err := c.client.SetVar(ctx, connect.NewRequest(&envsv1.SetVarRequest{Var: ev.ToProto()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}

func (c *client) RemoveVar(ctx context.Context, eid sdktypes.EnvID, vn sdktypes.Symbol) error {
	resp, err := c.client.RemoveVar(
		ctx,
		connect.NewRequest(
			&envsv1.RemoveVarRequest{
				EnvId: eid.String(),
				Name:  vn.String(),
			},
		),
	)
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	return internal.Validate(resp.Msg)
}

func (c *client) GetVars(ctx context.Context, vns []sdktypes.Symbol, eid sdktypes.EnvID) ([]sdktypes.EnvVar, error) {
	resp, err := c.client.GetVars(ctx, connect.NewRequest(
		&envsv1.GetVarsRequest{
			Names: kittehs.TransformToStrings(vns),
			EnvId: eid.String(),
		},
	))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Vars, sdktypes.StrictEnvVarFromProto)
}

func (c *client) RevealVar(ctx context.Context, eid sdktypes.EnvID, vn sdktypes.Symbol) (string, error) {
	resp, err := c.client.RevealVar(ctx, connect.NewRequest(
		&envsv1.RevealVarRequest{
			EnvId: eid.String(),
			Name:  vn.String(),
		},
	))
	if err != nil {
		return "", rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return "", err
	}

	return resp.Msg.Value, nil
}
