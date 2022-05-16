package manifest

import (
	"context"
)

type Action struct {
	Desc string
	Run  func(context.Context, *Env) (string, error)
}

type Actions []*Action
