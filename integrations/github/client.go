package github

import (
	"context"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	integrationName = "github"
)

var (
	integrationID = sdktypes.NewIntegrationIDFromName(integrationName)

	desc = common.Descriptor(integrationName, "GitHub", "/static/images/github.svg")
)

type integration struct {
	vars sdkservices.Vars
}

func New(cvars sdkservices.Vars) sdkservices.Integration {
	i := &integration{vars: cvars}
	return sdkintegrations.NewIntegration(
		desc, sdkmodule.New(), connStatus(i), connTest(i),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Using X".
func connStatus(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			zap.L().Error("failed to read connection vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(vars.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		switch at.Value() {
		case integrations.OAuth, integrations.OAuthDefault:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using GitHub app"), nil
		case integrations.PAT:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using PAT + webhook"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

func connTest(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			zap.L().Error("failed to read connection vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		var gh *github.Client
		if pat := vs.Get(vars.PAT); pat.IsValid() {
			gh = github.NewClient(nil).WithAuthToken(vs.GetValue(vars.PAT))
		} else {
			gh, err = newClientWithInstallJWT(vs)
			if err != nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}
		}

		// Make an API call that works for PAT & OAuth
		_, _, err = gh.RateLimit.Get(ctx)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}
