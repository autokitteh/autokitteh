package pythonrt

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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
	// Start from first { if it's not too far the start.
	i := strings.IndexByte(msg, '{')
	if i < 0 || i > 20 {
		i = 0
	}

	var record map[string]any
	if err := json.Unmarshal([]byte(msg[i:]), &record); err != nil {
		return nil
	}

	// Key (runner_id) be in sync with runner/log.py
	if record["runner_id"] != runnerID {
		return nil
	}

	return record
}

func cleanString(text string) string {
	text = strings.ToValidUTF8(text, "")

	for i, c := range text {
		// Assume chars from newline are OK
		if c >= '\n' {
			return text[i:]
		}

		if i > 20 {
			break
		}
	}

	return text
}

func (d LogDispatcher) Print(text string) error {
	// Docker logs may contain garbage characters at the start
	text = cleanString(text)
	record := getRunnerLog(d.runnerID, text)
	if record == nil {
		return d.print(d.ctx, d.runID, text)
	}

	d.logRecord(record)
	return nil
}

func (d LogDispatcher) logRecord(record map[string]any) {
	// If the type assertion fails, pyLevel will be "" which translates to Info level
	pyLevel, _ := record["level"].(string)
	level := pyLevelToZap(pyLevel)

	const msgKey = "message"
	msg, ok := record[msgKey].(string)
	if !ok {
		msg = "runner log"
	} else {
		delete(record, msgKey)
	}

	d.log.Log(level, msg, zap.Any("record", record))
}

func pyLevelToZap(level string) zapcore.Level {
	// Should be in sync with runner/log.py
	switch level {
	case "DEBUG":
		return zap.DebugLevel
	case "WARNING":
		return zap.WarnLevel
	case "ERROR":
		return zap.ErrorLevel
	}

	return zap.InfoLevel
}
