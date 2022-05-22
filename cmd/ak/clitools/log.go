package clitools

import (
	LL "github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/pkg/z"
)

var l LL.L = LL.Nop

func initLog(level string) error {
	if level == "" {
		level = "warn"
	}

	cfg := Z.Config{
		Level: level,
		Dev:   true,
	}

	var err error

	if l, err = Z.NewL(cfg, nil, nil); err != nil {
		return err
	}

	return nil
}

func L() LL.L { return l }

func Debugw(msg string, args ...interface{}) { l.Debug(msg, args...) }
func Debugf(msg string, args ...interface{}) { l.Debugf(msg, args...) }

func Infow(msg string, args ...interface{}) { l.Info(msg, args...) }
func Infof(msg string, args ...interface{}) { l.Infof(msg, args...) }

func Warnw(msg string, args ...interface{}) { l.Warn(msg, args...) }
func Warnf(msg string, args ...interface{}) { l.Warnf(msg, args...) }
