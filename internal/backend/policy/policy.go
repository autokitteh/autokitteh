package policy

import (
	"context"
)

// Make a decision based on the configured policy.
// `path` is the path to the value that the decision is being made on.
// `input` is additional context data for the decision making.
type DecideFunc func(ctx context.Context, path string, input any) (any, error)
