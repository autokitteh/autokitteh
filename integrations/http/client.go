package http

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var IntegrationID = sdktypes.IntegrationIDFromName("http")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: IntegrationID.String(),
	UniqueName:    "http",
	DisplayName:   "HTTP",
	LogoUrl:       "/static/images/http.svg",
	ConnectionUrl: "/i/http/connect",
}))

type integration struct {
	secrets sdkservices.Secrets
	scope   string
}

func New(sec sdkservices.Secrets) sdkservices.Integration {
	i := integration{secrets: sec, scope: desc.UniqueName().String()}
	return sdkintegrations.NewIntegration(desc, sdkmodule.New(
		sdkmodule.WithConfigAsData(),

		sdkmodule.ExportFunction(
			"get",
			i.request(http.MethodGet),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#GET"),
			args),
		sdkmodule.ExportFunction(
			"head",
			i.request(http.MethodHead),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#HEAD"),
			args),
		sdkmodule.ExportFunction(
			"post",
			i.request(http.MethodPost),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#POST"),
			args),
		sdkmodule.ExportFunction("put",
			i.request(http.MethodPut),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#PUT"),
			args),
		sdkmodule.ExportFunction(
			"delete",
			i.request(http.MethodDelete),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#DELETE"),
			args),
		sdkmodule.ExportFunction(
			"options",
			i.request(http.MethodOptions),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc9110#OPTIONS"),
			args),
		sdkmodule.ExportFunction(
			"patch",
			i.request(http.MethodPatch),
			sdkmodule.WithFuncDoc("https://www.rfc-editor.org/rfc/rfc5789"),
			args),
	))
}
