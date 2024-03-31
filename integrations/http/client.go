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
	"raw_body?",
	"form_body?",
	"json_body?",
	// TODO: Mismatched naming, see http.go lines 57-59.
	"form_encoding?",
	"auth?",
)

var IntegrationID = sdktypes.IntegrationIDFromName("http")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: IntegrationID.String(),
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
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#GET"),
			args),
		sdkmodule.ExportFunction(
			"head",
			request(http.MethodHead),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#HEAD"),
			args),
		sdkmodule.ExportFunction(
			"post",
			request(http.MethodPost),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#POST"),
			args),
		sdkmodule.ExportFunction("put",
			request(http.MethodPut),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#PUT"),
			args),
		sdkmodule.ExportFunction(
			"delete",
			request(http.MethodDelete),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#DELETE"),
			args),
		sdkmodule.ExportFunction(
			"options",
			request(http.MethodOptions),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#OPTIONS"),
			args),
		sdkmodule.ExportFunction(
			"patch",
			request(http.MethodPatch),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc5789"),
			args),
	))
}
