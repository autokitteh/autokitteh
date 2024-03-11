package grpc

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func New(desc sdktypes.Integration) sdkservices.Integration {
	return sdkintegrations.NewIntegration(desc, newGRPCModule)
}
