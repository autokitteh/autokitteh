package apply

import (
	"context"
)

type Operation struct {
	Description string                          `json:"description"`
	Action      func(ctx context.Context) error `json:"-"`
}

func (o Operation) String() string { return o.Description }
