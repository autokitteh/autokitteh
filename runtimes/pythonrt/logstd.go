package pythonrt

import (
	"bytes"
	"context"
	"sync"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type streamLogger struct {
	ctx    context.Context
	prefix string // prefix appended to each printed line
	print  sdkservices.RunPrintFunc
	rid    sdktypes.RunID

	mu  sync.Mutex
	buf bytes.Buffer
}

func newStreamLogger(ctx context.Context, prefix string, print sdkservices.RunPrintFunc, rid sdktypes.RunID) *streamLogger {
	s := streamLogger{
		prefix: prefix,
		print:  print,
		rid:    rid,
		ctx:    ctx,
	}
	s.reset()
	return &s
}

// reset resets the buffer to contain only `s.prefix`.
func (s *streamLogger) reset() {
	// No call to s.mu.Lock since it's called from Write which already acquires the lock.
	s.buf.Reset()
	s.buf.WriteString(s.prefix)
}

// Write implement io.Writer interface.
func (s *streamLogger) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data := p
	// io.Writer works with []byte, not lines.
	// `data` might contain several newlines (or none), so we iterate over it.
	for {
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			break
		}

		s.buf.Write(data[:i])
		s.print(s.ctx, s.rid, s.buf.String())

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
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.buf.Len() > 0 {
		s.print(context.Background(), s.rid, s.buf.String())
	}

	return nil
}
