package litterboxes

import (
	"embed"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed *
var FS embed.FS

type Example struct {
	Events map[string]string `json:"events"`
	Source string            `json:"source"`
}

var (
	Examples     = make(map[string]*Example)
	JSONExamples string
)

func init() {
	if err := fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == "." {
			return nil
		}

		if d.IsDir() {
			Examples[path] = &Example{
				Events: make(map[string]string),
			}
		} else if d := filepath.Dir(path); d != "." {
			x := Examples[d]
			if x == nil {
				panic("no such example")
			}

			bs, err := fs.ReadFile(FS, path)
			if err != nil {
				panic(err)
			}

			txt := string(bs)

			if filepath.Base(path) == "source" {
				x.Source = txt
			} else if filepath.Ext(path) == ".json" {
				x.Events[strings.TrimSuffix(filepath.Base(path), ".json")] = txt
			}

			return nil
		}

		return nil
	}); err != nil {
		panic(err)
	}

	bs, err := json.Marshal(Examples)
	if err != nil {
		panic(err)
	}

	JSONExamples = string(bs)
}
