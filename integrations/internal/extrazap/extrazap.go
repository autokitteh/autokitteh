// Package extrazap provides helper functions for initializing Zap loggers,
// as well as associating and extracting them with/from context objects.
package extrazap

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

// NewDevelopmentLogger builds a customized development logger which is even
// more user-friendly than [zap.NewDevelopment].
func NewDevelopmentLogger() *zap.Logger {
	c := zap.NewDevelopmentConfig()
	c.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	c.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	// TODO: Pretty-print tags like zerolog, instead of a JSON string?
	// TODO: Color "error" and "Error" tag values in red, like zerolog?
	// cfg.EncoderConfig.ConsoleSeparator = " "
	// TODO: Check out https://github.com/charmbracelet/log
	return zap.Must(c.Build())
}

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
