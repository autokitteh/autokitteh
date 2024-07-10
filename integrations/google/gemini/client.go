package gemini

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var integrationID = sdktypes.NewIntegrationIDFromName("googlegemini")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "googlegemini",
	DisplayName:   "Google Gemini",
	Description:   "Gemini is a generative artificial intelligence chatbot developed by Google.",
	LogoUrl:       "/static/images/google_gemini.svg",
	UserLinks: map[string]string{
		"1 REST API reference": "https://ai.google.dev/api/rest",
		"2 Python client API":  "https://ai.google.dev/api/python/google/generativeai",
		"3 Python samples":     "https://github.com/google-gemini/generative-ai-python/tree/main/samples",
	},
	ConnectionUrl: "/googlegemini/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New( /* No exported functions for Starlark */ ),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}
