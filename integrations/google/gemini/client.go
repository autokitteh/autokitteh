package gemini

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

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
		connTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// connTest is an optional connection test provided by the integration
// to AutoKitteh. It is used to verify that the connection is working
// as expected. The possible results are "OK" and "error".
func connTest(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		apiKey := vs.Get(apiKeyVar)
		if !apiKey.IsValid() || apiKey.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}
		url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + apiKey.Value()
		requestBody := `{
				"contents": [{
					"parts": [{"text": "Hello, Gemini!"}]
				}]
			}`

		req, err := http.NewRequest("POST", url, strings.NewReader(requestBody))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}
		defer resp.Body.Close()

		_, err = io.ReadAll(resp.Body)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		if resp.StatusCode != http.StatusOK {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, fmt.Sprintf("Request failed. Status Code: %d", resp.StatusCode)), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}
