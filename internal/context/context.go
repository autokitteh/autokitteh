package context

import (
	"context"
)

type ctxKey string

const componentCtxKey = ctxKey("component")

type ComponentType int

const (
	Dispatcher ComponentType = iota
	Workflow
	EventWorkflow
	SessionWorkflow
	Middleware
	Unknown
)

func (c ComponentType) String() string {
	if c >= Unknown {
		return "unknown"
	}
	return [...]string{"dispatcher", "workflow", "eventsWF", "sessionWF", "middleware"}[c]
}

func Component(ctx context.Context) ComponentType {
	if v := ctx.Value(componentCtxKey); v != nil {
		return v.(ComponentType)
	}
	return Unknown
}

func WithComponent(ctx context.Context, component ComponentType) context.Context {
	if component >= Unknown {
		return ctx
	}
	return context.WithValue(ctx, componentCtxKey, component)
}
