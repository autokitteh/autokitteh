package github

import (
	"github.com/google/go-github/v60/github"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var valueWrapper = sdktypes.ValueWrapper{
	Prewrap: func(v any) (any, error) {
		switch v := v.(type) {
		case github.Timestamp:
			return v.Time, nil
		}

		return v, nil
	},
}
