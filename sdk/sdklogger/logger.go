package sdklogger

import (
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

// zap.SugaredLogger style.
type Logger interface {
	Warn(...any)
	Error(...any)
	Panic(...any)
	DPanic(...any)
}

var logger Logger = defaultLogger{}

func SetGlobalLogger(l Logger) { logger = l }

func Warn(args ...any)   { logger.Warn(args...) }
func Error(args ...any)  { logger.Error(args...) }
func Panic(args ...any)  { logger.Panic(args...) }
func DPanic(args ...any) { logger.DPanic(args...) }
func DPanicOrReturn(args ...any) error {
	DPanic(args...)
	return errors.New(fmt.Sprint(args...))
}

// Default

type defaultLogger struct{}

func (defaultLogger) Warn(args ...any) {
	fmt.Printf("\nWARN: %v\n", fmt.Sprint(args...))
}

func (defaultLogger) Error(args ...any) {
	fmt.Printf("\nERROR: %v\n", fmt.Sprint(args...))
}

func (defaultLogger) Panic(args ...any) {
	// TODO: this could be better, for instance when args = {error}.
	kittehs.Panic(fmt.Errorf("%v", fmt.Sprint(args...)))
}

func (l defaultLogger) DPanic(args ...any) {
	l.Error(args...)
}
