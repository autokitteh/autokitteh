package grpc

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var integrationID = sdktypes.IntegrationIDFromName("grpc")

var description = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "grpc",
	DisplayName:   "GRPC",
	LogoUrl:       "/static/images/http.svg",
}))

func New(sec sdkservices.Secrets) sdkservices.Integration {
	return sdkintegrations.NewIntegration(description, sdkmodule.New(sdkmodule.ExportFunction("call", createGRPCCallWrapper("call"))))
}
