package slackeventsrcsvc

import (
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type logger struct{ L.L }

func (l logger) Output(_ int, msg string) error {
	l.L.Debug(msg)
	return nil
}

func wrapLogger(l L.L) logger { return logger{l} }
