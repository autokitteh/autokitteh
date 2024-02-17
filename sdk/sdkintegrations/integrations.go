package sdkintegrations

import (
	"context"
	"slices"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integrations map[string]sdkservices.Integration

var _ sdkservices.Integrations = (integrations)(nil)

func New(is []sdkservices.Integration) sdkservices.Integrations {
	return integrations(kittehs.ListToMap(is, func(i sdkservices.Integration) (string, sdkservices.Integration) {
		return sdktypes.GetIntegrationID(i.Get()).String(), i
	}))
}

func (is integrations) Get(ctx context.Context, id sdktypes.IntegrationID) (sdkservices.Integration, error) {
	return is[id.String()], nil
}

func (is integrations) List(ctx context.Context, nameSubstring string) ([]sdktypes.Integration, error) {
	// FIXME: Filter by nameSubstring (unique/display name).
	out := kittehs.FilterNils(kittehs.TransformMapToList(is, func(_ string, i sdkservices.Integration) sdktypes.Integration {
		desc := i.Get()
		return desc
	}))

	slices.SortFunc(out, func(a, b sdktypes.Integration) int {
		return strings.Compare(sdktypes.GetIntegrationDisplayName(a), sdktypes.GetIntegrationDisplayName(b))
	})

	return out, nil
}
