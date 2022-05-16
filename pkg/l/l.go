package l

import (
	"fmt"
	"strings"
)

type L interface {
	Named(string) L
	With(...interface{}) L

	SkipCaller(int) L

	Debug(string, ...interface{})
	Debugf(string, ...interface{})

	Info(string, ...interface{})
	Infof(string, ...interface{})

	Warn(string, ...interface{})
	Warnf(string, ...interface{})

	Error(string, ...interface{})
	Errorf(string, ...interface{})

	Fatal(string, ...interface{})
	Fatalf(string, ...interface{})

	Panic(string, ...interface{})
	Panicf(string, ...interface{})

	DPanic(string, ...interface{})
	DPanicf(string, ...interface{})
}

func Error(l L, msg string, pairs ...interface{}) error {
	N(l).Error(msg, pairs...)

	args := make([]string, 0, len(pairs)/2)

	for len(pairs) > 1 {
		args = append(args, fmt.Sprintf("%s: %v", pairs[0], pairs[1]))
	}

	if len(pairs) > 0 {
		N(l).DPanic("odd number of pairs", "pairs", pairs)
	}

	return fmt.Errorf("%s: %v", msg, strings.Join(args, ","))
}
