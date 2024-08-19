package github

import (
	"context"

	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (i integration) getCopilotBilling(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var org string

	if err := sdkmodule.UnpackArgs(args, kwargs, "org", &org); err != nil {
		return sdktypes.InvalidValue, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Copilot.GetCopilotBilling(ctx, org)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

func (i integration) listCopilotSeats(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		org  string
		opts github.ListOptions
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "org", &org, &opts); err != nil {
		return sdktypes.InvalidValue, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Copilot.ListCopilotSeats(ctx, org, &opts)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

func (i integration) addCopilotTeams(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		org   string
		teams []string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "org", &org, "teams", &teams); err != nil {
		return sdktypes.InvalidValue, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Copilot.AddCopilotTeams(ctx, org, teams)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

func (i integration) removeCopilotTeams(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		org   string
		teams []string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "org", &org, "teams", &teams); err != nil {
		return sdktypes.InvalidValue, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Copilot.RemoveCopilotTeams(ctx, org, teams)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

func (i integration) addCopilotUsers(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		org   string
		users []string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "org", &org, "users", &users); err != nil {
		return sdktypes.InvalidValue, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Copilot.AddCopilotUsers(ctx, org, users)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

func (i integration) removeCopilotUsers(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		org   string
		users []string
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "org", &org, "users", &users); err != nil {
		return sdktypes.InvalidValue, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Copilot.RemoveCopilotUsers(ctx, org, users)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}

func (i integration) getCopilotSeatDetails(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var org, user string

	if err := sdkmodule.UnpackArgs(args, kwargs, "org", &org, "user", &user); err != nil {
		return sdktypes.InvalidValue, err
	}

	gh, err := i.NewClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	resp, _, err := gh.Copilot.GetSeatDetails(ctx, org, user)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return valueWrapper.Wrap(resp)
}
