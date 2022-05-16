package tmplrender

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Masterminds/sprig/v3"
)

type RenderFunc func(http.ResponseWriter, string, interface{})

type Renderer struct {
	FS      fs.FS
	NoCache bool

	cache map[string]*template.Template
}

var _ RenderFunc = (&Renderer{}).Render

func New(path string, embedded fs.FS) *Renderer {
	fs := embedded

	if path != "" {
		fs = os.DirFS(path)
	}

	return &Renderer{FS: fs, NoCache: path != ""}
}

func (r *Renderer) parse(name string) *template.Template {
	return template.Must(
		template.New(filepath.Base(name)).
			Funcs(sprig.FuncMap()).
			Funcs(funcMap).
			ParseFS(r.FS, "*"),
	)
}

func (r *Renderer) Render(w http.ResponseWriter, name string, ctx interface{}) {
	var t *template.Template

	if r.NoCache {
		t = r.parse(name)
	} else {
		if r.cache == nil {
			r.cache = make(map[string]*template.Template)
		}

		if t = r.cache[name]; t == nil {
			t = r.parse(name)
			r.cache[name] = t
		}
	}

	if t == nil {
		panic(fmt.Errorf("cannot load template %q", name))
	}

	if err := t.Execute(w, ctx); err != nil {
		http.Error(w, fmt.Sprintf("render: %v", err), http.StatusInternalServerError)
	}
}
