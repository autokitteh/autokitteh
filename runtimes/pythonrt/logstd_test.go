package pythonrt

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var scriptTemplate = `#!/bin/bash

echo '%s'
echo '%s' >&2
`

func Test_streamLogger(t *testing.T) {
	outMsg := "info"
	errMsg := "error"
	exe := path.Join(t.TempDir(), "print.sh")
	file, err := os.Create(exe)
	require.NoError(t, err)
	fmt.Fprintf(file, scriptTemplate, outMsg, errMsg)
	file.Close()
	err = os.Chmod(exe, 0755)
	require.NoError(t, err)

	var buf bytes.Buffer

	print := func(ctx context.Context, rid sdktypes.RunID, text string) {
		buf.WriteString(text)
		buf.WriteString("\n")
	}
	rid := sdktypes.NewRunID()

	outPrefix, errPrefix := "[stdout] ", "[stderr] "
	stdout := newStreamLogger(outPrefix, print, rid)
	stderr := newStreamLogger(errPrefix, print, rid)

	cmd := exec.Command(exe)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	require.NoError(t, err)
	stdout.Close()
	stderr.Close()

	out := buf.String()
	require.Contains(t, out, fmt.Sprintf("%s%s", outPrefix, outMsg))
	require.Contains(t, out, fmt.Sprintf("%s%s", errPrefix, errMsg))
}

func Test_streamLogger_MultiLine(t *testing.T) {
	var buf bytes.Buffer

	print := func(ctx context.Context, rid sdktypes.RunID, text string) {
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	stdout := newStreamLogger("[stdout] ", print, sdktypes.NewRunID())
	stdout.Write([]byte("garfield\ngrumpy\npuss"))
	stdout.Close()

	fields := strings.Split(strings.TrimSpace(buf.String()), "\n")
	expected := []string{
		"[stdout] garfield",
		"[stdout] grumpy",
		"[stdout] puss",
	}
	require.Equal(t, expected, fields)
}
