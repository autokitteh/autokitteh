package sdkintegrations

import (
	"context"
	"slices"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integrations map[sdktypes.IntegrationID]sdkservices.Integration

var _ sdkservices.Integrations = (integrations)(nil)

func New(is []sdkservices.Integration) sdkservices.Integrations {
	return integrations(kittehs.ListToMap(is, func(i sdkservices.Integration) (sdktypes.IntegrationID, sdkservices.Integration) {
		return i.Get().ID(), i
	}))
}

func (is integrations) Get(ctx context.Context, id sdktypes.IntegrationID) (sdkservices.Integration, error) {
	return is[id], nil
}

func (is integrations) List(ctx context.Context, nameSubstring string) ([]sdktypes.Integration, error) {
	// FIXME: Filter by nameSubstring (unique/display name).
	out := kittehs.Filter(kittehs.TransformMapToList(is, func(_ sdktypes.IntegrationID, i sdkservices.Integration) sdktypes.Integration {
		return i.Get()
	}), sdktypes.IsValid)

	slices.SortFunc(out, func(a, b sdktypes.Integration) int {
		return strings.Compare(a.DisplayName(), b.DisplayName())
	})

	return out, nil
}
