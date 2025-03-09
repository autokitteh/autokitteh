package jira

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	integrationName = "jira"
)

var (
	IntegrationID = sdktypes.NewIntegrationIDFromName(integrationName)

	desc = common.Descriptor(integrationName, "Atlassian Jira", "/static/images/jira.svg")
)

type integration struct {
	vars sdkservices.Vars
}

func New(cvars sdkservices.Vars) sdkservices.Integration {
	i := &integration{vars: cvars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		connStatus(i),
		connTest(i),
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

		at := vs.Get(authType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		switch at.Value() {
		case integrations.APIToken:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using cloud API token"), nil
		case integrations.OAuth:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using OAuth 2.0"), nil
		case integrations.PAT:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using on-prem PAT"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// connTest is an optional connection test provided by the integration
// to AutoKitteh. It is used to verify that the connection is working
// as expected. The possible results are "OK" and "error".
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

		at := vs.Get(authType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		switch at.Value() {
		case integrations.OAuth:
			err := oauthConnTest(vs)
			if err != nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}
		case integrations.APIToken:
			err := apiTokenConnTest(ctx, nil, vs)
			if err != nil {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
			}
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}

// oauthConnTest verifies the connection for OAuth authentication.
// It checks the accessible resources using the OAuth access token.
func oauthConnTest(vs sdktypes.Vars) error {
	baseURL, err := APIBaseURL()
	if err != nil {
		return err
	}

	// TODO(INT-173): Create & use a new access token using the refresh token.
	token := vs.GetValueByString("oauth_AccessToken")
	_, err = accessibleResources(context.Background(), nil, baseURL, token)

	return err
}

// apiTokenConnTest verifies the connection for API key authentication.
// It sends a request to the API to confirm credentials and access.
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-myself/#api-group-myself
func apiTokenConnTest(ctx context.Context, l *zap.Logger, vs sdktypes.Vars) error {
	ctx, cancel := context.WithTimeout(ctx, common.HTTPTimeout)
	defer cancel()

	baseURL := vs.Get(baseURL).Value()
	email := vs.Get(email).Value()
	token := vs.Get(token).Value()

	u := baseURL + "/rest/api/3/myself"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		logWarnIfNotNil(l, "Failed to construct HTTP request for Jira API token test", zap.Error(err))
		return err
	}

	req.SetBasicAuth(email, token)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logWarnIfNotNil(l, "Failed to request current user info for Jira API token", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logWarnIfNotNil(l, "Unexpected response on current user info", zap.Int("status", resp.StatusCode))
		return fmt.Errorf("current user info: unexpected status code %d", resp.StatusCode)
	}

	return nil
}
