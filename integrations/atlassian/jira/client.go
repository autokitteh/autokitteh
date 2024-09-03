package jira

import (
	"context"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct {
	vars sdkservices.Vars
}

var integrationID = sdktypes.NewIntegrationIDFromName("jira")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "jira",
	DisplayName:   "Atlassian Jira",
	Description:   "Atlassian Jira is an issue tracking and project management system.",
	LogoUrl:       "/static/images/jira.svg",
	UserLinks: map[string]string{
		"1 REST API":                    "https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/",
		"2 Atlassian Python client API": "https://atlassian-python-api.readthedocs.io/",
		"3 Atlassian Python examples":   "https://github.com/atlassian-api/atlassian-python-api/tree/master/examples/jira",
		"4 Jira Python client API":      "https://jira.readthedocs.io/",
		"5 Jira Python examples":        "https://github.com/pycontribs/jira/tree/main/examples",
	},
	ConnectionUrl: "/jira/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	i := &integration{vars: cvars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New( /* No exported functions for Starlark */ ),
		connStatus(i),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "init
// required" (the connection is not usable yet) and "using X".
func connStatus(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(authType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		switch at.Value() {
		case integrations.APIToken:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using cloud API token"), nil
		case integrations.OAuth:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using OAuth 2.0"), nil
		case integrations.PAT:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using on-prem PAT"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "bad auth type"), nil
		}
	})
}
