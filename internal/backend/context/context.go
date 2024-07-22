package context

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ctxKey string

const componentCtxKey = ctxKey("origin")

type RequestOrginatorType int

const (
	Dispatcher RequestOrginatorType = iota
	EventWorkflow
	SessionWorkflow
	ScheduleWorkflow
	User
	Unknown
)

func (c RequestOrginatorType) String() string {
	if c >= Unknown {
		return "unknown"
	}
	return [...]string{"dispatcher", "eventsWorkflow", "sessionWorkflow", "schedulerWorkflow", "user"}[c]
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

func WithOwnershipOf(ctx context.Context, entityOwnership func(context.Context, sdktypes.UUID) (string, error), entityID sdktypes.UUID) context.Context {
	uid, err := entityOwnership(ctx, entityID)
	if err != nil {
		return ctx
	}
	return authcontext.SetAuthnUserID(ctx, uid)
}
