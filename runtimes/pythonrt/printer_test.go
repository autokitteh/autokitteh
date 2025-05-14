package pythonrt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func ExamplePrinter() {
	printFn := func(line string) error {
		fmt.Println(line)
		return nil
	}

	printer := Printer{
		printFn: printFn,
	}

	data := []byte("a\nb\nc\nd")
	if _, err := printer.Write(data); err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	printer.Flush()

	// Output:
	// a
	// b
	// c
	// d
}

func TestDispatcher(t *testing.T) {
	// Create a logger that captures logs
	core, logs := observer.New(zap.InfoLevel)
	log := zaptest.NewLogger(t, zaptest.WrapOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return core
	})))

	var buf bytes.Buffer

	d := LogDispatcher{
		ctx:      context.Background(),
		runnerID: "r1",
		runID:    sdktypes.NewRunID(),
		log:      log,
		print: func(ctx context.Context, rid sdktypes.RunID, text string) error {
			fmt.Fprint(&buf, text)
			return nil
		},
	}

	// session
	d.Print("Garfield\n")

	// operational
	msg := map[string]any{
		"runner_id": d.runnerID,
		"message":   "Grumpy",
	}
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	d.Print(string(data))

	// wrong runner, session
	msg["runner_id"] = d.runnerID + "z"
	data, err = json.Marshal(msg)
	require.NoError(t, err)
	err = d.Print(string(data))
	require.NoError(t, err)

	err = log.Sync()
	require.NoError(t, err)

	lines := strings.Split(buf.String(), "\n")
	require.Equal(t, 2, len(lines))
	require.Equal(t, 1, logs.Len())
}
