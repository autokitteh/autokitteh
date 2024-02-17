package http

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var args = sdkmodule.WithArgs(
	"url",
	"params?",
	"headers?",
	"body?",
	"form_body?",
	"form_encoding?",
	"json_body?",
	"auth?",
)

var integrationID = sdktypes.IntegrationIDFromName("http")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "http",
	DisplayName:   "HTTP",
	LogoUrl:       "/static/images/http.svg",
	// TODO: Integration documentation link.
}))

func New(sec sdkservices.Secrets) sdkservices.Integration {
	return sdkintegrations.NewIntegration(desc, sdkmodule.New(
		sdkmodule.ExportFunction(
			"get",
			request(http.MethodGet),
			sdkmodule.WithFuncDoc("https://datatracker.ietf.org/doc/html/rfc9110#name-get"),
			args),
		sdkmodule.ExportFunction(
			"head",
			request(http.MethodHead),
			sdkmodule.WithFuncDoc("https://datatracker.ietf.org/doc/html/rfc9110#name-head"),
			args),
		sdkmodule.ExportFunction(
			"post",
			request(http.MethodPost),
			sdkmodule.WithFuncDoc("https://datatracker.ietf.org/doc/html/rfc9110#name-post"),
			args),
		sdkmodule.ExportFunction("put",
			request(http.MethodPut),
			sdkmodule.WithFuncDoc("https://datatracker.ietf.org/doc/html/rfc9110#name-put"),
			args),
		sdkmodule.ExportFunction(
			"delete",
			request(http.MethodDelete),
			sdkmodule.WithFuncDoc("https://datatracker.ietf.org/doc/html/rfc9110#name-delete"),
			args),
		sdkmodule.ExportFunction(
			"options",
			request(http.MethodOptions),
			sdkmodule.WithFuncDoc("https://datatracker.ietf.org/doc/html/rfc9110#name-options"),
			args),
		sdkmodule.ExportFunction(
			"patch",
			request(http.MethodPatch),
			sdkmodule.WithFuncDoc("https://datatracker.ietf.org/doc/html/rfc5789"),
			args),
	))
}
