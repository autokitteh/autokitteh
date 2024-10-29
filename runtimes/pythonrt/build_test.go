package pythonrt

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.uber.org/zap"
)

func TestBuildFiles(t *testing.T) {
	m := map[string][]byte{
		"autokitteh.yaml":         []byte("ak"),
		"program.py":              []byte("py"),
		".gitignore":              []byte("git"),
		"__pycache__/program.pyc": []byte("pyc"),
	}

	mfs, err := kittehs.MapToMemFS(m)
	require.NoError(t, err)

	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	py := pySvc{
		log: log,
	}

	ba, err := py.Build(context.Background(), mfs, ".", nil)
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

	expected := []string{"autokitteh.yaml", "program.py"}
	require.Equal(t, expected, names)
}
