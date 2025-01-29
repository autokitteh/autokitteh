package nodejsrt

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

func TestBuildFiles(t *testing.T) {
	m := map[string][]byte{
		"autokitteh.yaml":         []byte("ak"),
		"program.js":              []byte("js"),
		".gitignore":              []byte("git"),
		"__pycache__/program.pyc": []byte("pyc"),
	}

	mfs, err := kittehs.MapToMemFS(m)
	require.NoError(t, err)

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	js := nodejsSvc{
		log: log,
	}

	ba, err := js.Build(context.Background(), mfs, ".", nil)
	require.NoError(t, err)

	tarData, ok := ba.CompiledData()[archiveKey]
	require.True(t, ok)
	tf := tar.NewReader(bytes.NewReader(tarData))

	var names []string
	for {
		hdr, err := tf.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		names = append(names, hdr.Name)
	}

	expected := []string{"autokitteh.yaml", "program.js"}
	require.Equal(t, expected, names)
}
