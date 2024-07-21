package context

import (
	"context"
)

type ctxKey string

const componentCtxKey = ctxKey("origin")

type RequestOrginatorType int

const (
	Dispatcher RequestOrginatorType = iota
	EventWorkflow
	SessionWorkflow
	ScheduleWorkflow
	Middleware
	Unknown
)

func (c RequestOrginatorType) String() string {
	if c >= Unknown {
		return "unknown"
	}
	return [...]string{"dispatcher", "eventsWorkflow", "sessionWorkflow", "schedulerWorkflow", "authMiddleware"}[c]
}

func RequestOrginator(ctx context.Context) RequestOrginatorType {
	if v := ctx.Value(componentCtxKey); v != nil {
		return v.(RequestOrginatorType)
	}
	return Unknown
}

func WithRequestOrginator(ctx context.Context, component RequestOrginatorType) context.Context {
	if component >= Unknown {
		return ctx
	}
	return context.WithValue(ctx, componentCtxKey, component)
}
