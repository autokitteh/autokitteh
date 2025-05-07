package pythonrt

import (
	"bytes"
	"context"
	"encoding/json"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

// Printer implements io.Writer and calls printFn for each line.
type Printer struct {
	buf     []byte
	printFn func(string) error
}

func (p *Printer) Write(b []byte) (int, error) {
	p.buf = append(p.buf, b...)
	offset := 0
	for {
		i := bytes.IndexByte(p.buf[offset:], '\n')
		if i == -1 {
			break
		}

		line := string(p.buf[offset : offset+i])
		if err := p.printFn(line); err != nil {
			return 0, err
		}
		offset += i + 1
	}

	copy(p.buf, p.buf[offset:])
	p.buf = p.buf[:len(p.buf)-offset]

	return len(b), nil
}

func (p *Printer) Flush() error {
	if len(p.buf) > 0 {
		if err := p.printFn(string(p.buf)); err != nil {
			return err
		}
		p.buf = p.buf[:0]
	}
	return nil
}

type LogDispatcher struct {
	ctx      context.Context
	runnerID string
	runID    sdktypes.RunID
	log      *zap.Logger
	print    func(ctx context.Context, rid sdktypes.RunID, text string) error
}

func getRunnerLog(runnerID, msg string) map[string]any {
	var record map[string]any
	if err := json.Unmarshal([]byte(msg), &record); err != nil {
		return nil
	}

	// Must be in sync with runner/log.py
	if record["runner_id"] != runnerID {
		return nil
	}

	return record
}

func (d LogDispatcher) Print(text string) error {
	record := getRunnerLog(d.runnerID, text)
	if record != nil {
		return d.print(d.ctx, d.runID, text)
	}

	d.log.Info("runner log", zap.Any("record", record))
	return nil
}
