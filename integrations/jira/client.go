package jira

import (
	"context"
	"fmt"
	"strings"

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
		"1 REST API":          "https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/",
		"2 Go client API":     "https://pkg.go.dev/github.com/andygrunwald/go-jira",
		"3 Python client API": "https://jira.readthedocs.io/",
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
		connTest(i),
	)
}

func connStatus(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		initReq := sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required")

		if !cid.IsValid() {
			return initReq, nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		// TODO(ENG-965): Implement a real check, and reuse in the OAuth handler.
		// if vs.Has(vars.PAT) {
		// 	return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using PAT"), nil
		// }

		n := len(kittehs.Filter(vs, func(v sdktypes.Var) bool {
			return strings.HasPrefix(v.Name().String(), "app_id__")
		}))

		if n == 0 {
			return initReq, nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, fmt.Sprintf("%d installations", n)), nil
	})
}

func connTest(*integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		// TODO
		return sdktypes.NewStatus(sdktypes.StatusCodeUnspecified, `¯\_(ツ)_/¯`), nil
	})
}
