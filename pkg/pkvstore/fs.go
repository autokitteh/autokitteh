package pkvstore

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

var _ Store = &FSStore{}

func (f *FSStore) pPath(p string) string {
	if enc, ok := f.Options["encode-p"]; ok {
		switch enc {
		case "":
			// nop
		case "hex":
			p = hex.EncodeToString([]byte(p))
		default:
			panic(fmt.Sprintf("unrecognized p encoder: %q", enc))
		}
	}

	return filepath.Join(f.RootPath, p)
}

func (f *FSStore) kPath(p, k string) string {
	p = f.pPath(p)

	if enc, ok := f.Options["encode-key"]; ok {
		switch enc {
		case "":
			// nop
		case "hex":
			k = hex.EncodeToString([]byte(k))
		default:
			panic(fmt.Sprintf("unrecognized k encoder: %q", enc))
		}
	}

	return filepath.Join(p, k)
}

func (f *FSStore) Put(_ context.Context, p, k string, v []byte) error {
	if p == "" || k == "" {
		return fmt.Errorf("either p or k must not be empty")
	}

	path := f.kPath(p, k)
	dir := f.pPath(p)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, fs.ModePerm); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
	}

	if err := ioutil.WriteFile(path, v, 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func (f *FSStore) Get(_ context.Context, p, k string) ([]byte, error) {
	path := f.kPath(p, k)

	v, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("read: %w", err)
	}

	return v, err
}

func (f *FSStore) Delete(_ context.Context, p, k string) error {
	path := f.kPath(p, k)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}

		return fmt.Errorf("rm: %w", err)
	}

	return nil
}

func (f *FSStore) List(_ context.Context, p string) ([]string, error) {
	path := f.pPath(p)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	ks := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		k := file.Name()

		if enc, ok := f.Options["encode-key"]; ok {
			switch enc {
			case "":
				// nop
			case "hex":
				dk, err := hex.DecodeString(k)
				if err != nil {
					return nil, fmt.Errorf("%s: %w", k, err)
				}

				k = string(dk)
			default:
				panic(fmt.Sprintf("unrecognized k encoder: %q", enc))
			}
		}

		ks = append(ks, k)
	}

	return ks, nil
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
