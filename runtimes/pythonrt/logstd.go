package pythonrt

import (
	"bytes"
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type streamLogger struct {
	prefix string // prefix appended to each printed line
	print  sdkservices.RunPrintFunc
	rid    sdktypes.RunID
	buf    bytes.Buffer
}

func newStreamLogger(prefix string, print sdkservices.RunPrintFunc, rid sdktypes.RunID) *streamLogger {
	s := streamLogger{
		prefix: prefix,
		print:  print,
		rid:    rid,
	}
	s.reset()
	return &s
}

// reset resets the buffer to contain only `s.prefix`.
func (s *streamLogger) reset() {
	s.buf.Reset()
	s.buf.WriteString(s.prefix)
}

// Write implement io.Writer interface.
func (s *streamLogger) Write(p []byte) (int, error) {
	ctx := context.Background()
	data := p
	// io.Writer works with []byte, not lines.
	// `data` might contain several newlines (or none), so we iterate over it.
	for {
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			break
		}

		s.buf.Write(data[:i])
		s.print(ctx, s.rid, s.buf.String())

		s.reset()
		data = data[i+1:]
	}

	if len(data) > 0 {
		s.buf.Write(data)
	}

	return len(p), nil
}

// Close implements io.Closer.
// It will print whatever in `s.buf`.
func (s *streamLogger) Close() error {
	if s.buf.Len() > 0 {
		s.print(context.Background(), s.rid, s.buf.String())
	}

	return nil
}
