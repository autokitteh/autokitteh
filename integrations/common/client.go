package common

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Descriptor(uniqueName, displayName, logoURL string) sdktypes.Integration {
	return kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
		IntegrationId: sdktypes.NewIntegrationIDFromName(uniqueName).String(),
		UniqueName:    uniqueName,
		DisplayName:   displayName,
		LogoUrl:       logoURL,
		ConnectionUrl: "/" + uniqueName,
		ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
			RequiresConnectionInit: true,
			SupportsConnectionTest: true,
		},
	}))
}

func LegacyDescriptor(uniqueName, displayName, logoURL string) sdktypes.Integration {
	u := fmt.Sprintf("/%s/connect", uniqueName)
	return Descriptor(uniqueName, displayName, logoURL).WithConnectionURL(u)
}
