package slackeventsrcsvc

import (
	L "gitlab.com/softkitteh/autokitteh/pkg/l"
)

type logger struct{ L.L }

func (l logger) Output(_ int, msg string) error {
	l.L.Debug(msg)
	return nil
}

func wrapLogger(l L.L) logger { return logger{l} }
