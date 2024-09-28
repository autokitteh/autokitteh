// Package extrazap provides helper functions for initializing Zap loggers,
// as well as associating and extracting them with/from context objects.
package extrazap

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

// AttachLoggerToContext returns a copy of the given context with the given logger
// attached to it. Neither the given context nor the given logger are affected.
func AttachLoggerToContext(l *zap.Logger, ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// ExtractLoggerFromContext returns the logger attached to the given context,
// or the global logger if no logger instance was previously attached to it.
func ExtractLoggerFromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return l
	}
	return zap.L()
}
