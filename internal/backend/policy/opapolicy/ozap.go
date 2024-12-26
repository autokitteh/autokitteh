// This is an adaptation of https://github.com/open-policy-agent/contrib/blob/main/logging/plugins/ozap/zap.go.
// For some reason, the original does not actually make use of wrapper.level. This fixes it.
package opapolicy

import (
	"fmt"

	"github.com/open-policy-agent/opa/v1/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Wrap this zap Logger and AtomicLevel as a logging.Logger.
func wrapLogger(log *zap.Logger, level *zap.AtomicLevel) logging.Logger {
	return &wrapper{internal: log, level: level}
}

// wrapper implements logging.Logger for a zap Logger.
type wrapper struct {
	internal *zap.Logger
	level    *zap.AtomicLevel
}

func (w *wrapper) shouldEmit(l zapcore.Level) bool {
	return w.level == nil || w.level.Enabled(l)
}

// Debug logs at debug level.
func (w *wrapper) Debug(f string, a ...interface{}) {
	if w.shouldEmit(zap.DebugLevel) {
		w.internal.Debug(fmt.Sprintf(f, a...))
	}
}

// Info logs at info level.
func (w *wrapper) Info(f string, a ...interface{}) {
	if w.shouldEmit(zap.InfoLevel) {
		w.internal.Info(fmt.Sprintf(f, a...))
	}
}

// Error logs at error level.
func (w *wrapper) Error(f string, a ...interface{}) {
	if w.shouldEmit(zap.ErrorLevel) {
		w.internal.Error(fmt.Sprintf(f, a...))
	}
}

// Warn logs at warn level.
func (w *wrapper) Warn(f string, a ...interface{}) {
	if w.shouldEmit(zap.WarnLevel) {
		w.internal.Warn(fmt.Sprintf(f, a...))
	}
}

// WithFields provides additional fields to include in log output.
func (w *wrapper) WithFields(fields map[string]interface{}) logging.Logger {
	return &wrapper{
		internal: w.internal.With(toZapFields(fields)...),
		level:    w.level,
	}
}

// toZapFields converts logging format fields to zap format Fields
func toZapFields(fields map[string]interface{}) []zap.Field {
	var zapFields []zap.Field
	for k, v := range fields {
		switch t := v.(type) {
		case error:
			zapFields = append(zapFields, zap.NamedError(k, t))
		case string:
			zapFields = append(zapFields, zap.String(k, t))
		case bool:
			zapFields = append(zapFields, zap.Bool(k, t))
		case int:
			zapFields = append(zapFields, zap.Int(k, t))
		default:
			zapFields = append(zapFields, zap.Any(k, v))
		}
	}
	return zapFields
}

// SetLevel sets the logger level.
func (w *wrapper) GetLevel() logging.Level {
	switch w.internal.Level() {
	case zap.ErrorLevel:
		return logging.Error
	case zap.WarnLevel:
		return logging.Warn
	case zap.DebugLevel:
		return logging.Debug
	default:
		return logging.Info
	}
}

// SetLevel sets the logger level.
func (w *wrapper) SetLevel(l logging.Level) {
	var newLevel zapcore.Level
	switch l {
	case logging.Error:
		newLevel = zap.ErrorLevel
	case logging.Warn:
		newLevel = zap.WarnLevel
	case logging.Info:
		newLevel = zap.InfoLevel
	case logging.Debug:
		newLevel = zap.DebugLevel
	default:
		return
	}
	w.level.SetLevel(newLevel)
}
