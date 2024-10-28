package kittehs

import (
	"io/fs"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterFS(t *testing.T) {

	m := map[string][]byte{
		"autokitteh.yaml": []byte("ak"),
		"program.py":      []byte("py"),
		".gitignore":      []byte("git"),
	}

	mfs, err := MapToMemFS(m)
	require.NoError(t, err)

	ffs, err := NewFilterFS(mfs, nil) // no pred
	require.NoError(t, err)
	matches, err := fs.Glob(ffs, "*")
	require.NoError(t, err)
	files := make([]string, 0, len(m))
	for name := range m {
		files = append(files, name)
	}
	require.Equal(t, files, matches)

	// Ignore files starting with .
	isOK := func(e fs.DirEntry) bool {
		return path.Base(e.Name())[0] != '.'
	}

	ffs, err = NewFilterFS(mfs, isOK) // no pred
	require.NoError(t, err)
	matches, err = fs.Glob(ffs, "*")
	require.NoError(t, err)
	files = Filter(files, func(name string) bool { return name[0] != '.' })
	require.Equal(t, files, matches)
}
