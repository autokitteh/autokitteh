package azurebot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/azurebot/webhooks"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var desc = common.Descriptor("azurebot", "Azure Bot Service", "/static/images/azure_bot.svg")

type integration struct {
	vars sdkservices.Vars
	l    *zap.Logger
}

func New(vars sdkservices.Vars, l *zap.Logger) sdkservices.Integration {
	i := &integration{vars: vars, l: l}

	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		i.connStatus(),
		i.connTest(),
		sdkintegrations.WithConnectionConfigFromVars(vars),
	)
}

func (i *integration) connStatus() sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		l := i.l.With(zap.String("connection_id", cid.String()))

		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("failed to read connection "+cid.String()+" vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		var vars webhooks.Vars
		vs.Decode(&vars)

		if vars.AppID != "" && vars.AppPassword != "" && vars.TenantID != "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Not initialized"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Initialized"), nil
	})
}

func (i *integration) connTest() sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		l := i.l.With(zap.String("connection_id", cid.String()))

		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("failed to read connection "+cid.String()+" vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		var vars webhooks.Vars
		vs.Decode(&vars)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://login.microsoftonline.com/botframework.com/oauth2/v2.0/token", nil)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "failed to create request"), nil
		}

		req.PostForm = url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {vars.AppID},
			"client_secret": {vars.AppPassword},
			"scope":         {"https://api.botframework.com/.default"},
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			l.Debug("test: failed to connect to token endpoint", zap.Error(err))
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "failed to connect"), nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			l.Debug("test: failed to read response body", zap.Error(err))
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "failed to read response body"), nil
		}

		if resp.StatusCode != http.StatusOK {
			l.Debug("test: non-200 response from token endpoint", zap.String("status", resp.Status), zap.ByteString("body", body))
			return sdktypes.NewStatusf(sdktypes.StatusCodeError, "error: %s", resp.Status), nil
		}

		var data struct {
			Error       string `json:"error"`
			AccessToken string `json:"access_token"`
		}

		if err := json.Unmarshal(body, &data); err != nil {
			l.Debug("test: failed to parse response body", zap.Error(err), zap.ByteString("body", body))
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "failed to parse response body"), nil
		}

		if data.Error != "" {
			l.Debug("test: error response from token endpoint", zap.String("error", data.Error), zap.ByteString("body", body))
			return sdktypes.NewStatus(sdktypes.StatusCodeError, data.Error), nil
		}

		if data.AccessToken == "" {
			l.Debug("test: missing access token in response", zap.ByteString("body", body))
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "missing access token"), nil
		}

		l.Debug("test: successfully obtained access token")
		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Authenticated"), nil
	})
}
