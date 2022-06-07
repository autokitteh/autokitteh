package litterboxes

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed *
var FS embed.FS

type Example struct {
	Events  map[string]string `json:"events"`
	Program string            `json:"program"`
}

var (
	Examples     = make(map[string]*Example)
	JSONExamples string
)

func init() {
	dis, err := FS.ReadDir(".")
	if err != nil {
		panic(err)
	}

	for _, di := range dis {
		if !di.IsDir() {
			continue
		}

		dname := filepath.Base(di.Name())

		dis, err := FS.ReadDir(di.Name())
		if err != nil {
			panic(err)
		}

		for _, di := range dis {
			if di.IsDir() {
				continue
			}

			name := filepath.Base(di.Name())

			prep := func(sub string) *Example {
				key := dname
				if sub != "" {
					key = fmt.Sprintf("%s/%s", dname, sub)
				}

				x := Examples[key]
				if x == nil {
					x = &Example{Events: make(map[string]string)}
					Examples[key] = x
				}

				return x
			}

			a, b, _ := strings.Cut(name, "--")

			var (
				x    *Example
				kind string
			)

			if b == "" {
				x = prep("")
				kind = a
			} else {
				x = prep(a)
				kind = b
			}

			bs, err := fs.ReadFile(FS, filepath.Join(dname, di.Name()))
			if err != nil {
				panic(err)
			}

			txt := string(bs)

			if strings.HasSuffix(kind, ".json") {
				n := strings.TrimSuffix(kind, ".json")

				x.Events[n] = txt
			} else if kind == "program" {
				x.Program = txt
			}
		}
	}

	bs, err := json.Marshal(Examples)
	if err != nil {
		panic(err)
	}

	JSONExamples = string(bs)
}
