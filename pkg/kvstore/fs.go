package kvstore

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FSStore struct {
	RootPath string
	Options  map[string]string
}

func (f *FSStore) keyPath(k string) string {
	if enc, ok := f.Options["encode-key"]; ok {
		switch enc {
		case "":
			// nop
		case "hex":
			k = hex.EncodeToString([]byte(k))
		default:
			panic(fmt.Sprintf("unrecognized key encoder: %q", enc))
		}
	}

	return filepath.Join(f.RootPath, k)
}

func (f *FSStore) Put(_ context.Context, k string, v []byte) error {
	if k == "" {
		return fmt.Errorf("k must not be empty")
	}

	path := f.keyPath(k)

	if err := ioutil.WriteFile(path, v, 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func (f *FSStore) Get(_ context.Context, k string) ([]byte, error) {
	path := f.keyPath(k)

	v, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("read: %w", err)
	}

	return v, err
}

func (f *FSStore) Delete(_ context.Context, k string) error {
	path := f.keyPath(k)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}

		return fmt.Errorf("rm: %w", err)
	}

	return nil
}

func (f *FSStore) Setup(context.Context) error {
	if err := os.MkdirAll(f.RootPath, fs.ModePerm); err != nil {
		return fmt.Errorf("mkdirall: %w", err)
	}

	return nil
}

func (f *FSStore) Teardown(context.Context) error {
	if err := os.RemoveAll(f.RootPath); err != nil {
		return fmt.Errorf("removeall: %w", err)
	}

	return nil
}
