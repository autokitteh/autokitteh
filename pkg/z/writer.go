package z

import (
	"io"

	"github.com/Songmu/axslogparser"
	"go.uber.org/zap"
)

// ApacheLogWriter implements an io.Writer interface to be
// used as a sink for logging.
type ApacheLogWriter struct {
	Z *zap.SugaredLogger

	// True will report regular access as Info, False as Debug.
	InfoLevel bool
}

var _ io.Writer = &ApacheLogWriter{}

var parser axslogparser.Apache

func (l *ApacheLogWriter) Write(bs []byte) (int, error) {
	if l.Z == nil {
		return len(bs), nil
	}

	log, err := parser.Parse(string(bs))
	if err != nil {
		l.Z.Warnw("combined log parse error", "err", err, "raw", string(bs))
	} else if log.Status >= 500 {
		l.Z.Errorw("http", "log", log)
	} else if l.InfoLevel {
		l.Z.Infow("http", "log", log)
	} else {
		l.Z.Debugw("http", "log", log)
	}

	return len(bs), nil
}
