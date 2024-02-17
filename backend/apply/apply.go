package apply

import (
	"context"
	"fmt"
)

func (a *Applicator) Apply(ctx context.Context) error {
	for i, op := range a.ops {
		if err := op.Action(ctx); err != nil {
			return fmt.Errorf("apply %d %q: %w", i, op.Description, err)
		}
	}

	return nil
}
