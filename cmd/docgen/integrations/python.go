package integrations

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
)

type pythonGenerator struct {
	outputDir string
}

func (g pythonGenerator) Output() string {
	return fmt.Sprintf("Python stubs (%s)", g.outputDir)
}

func NewPythonGenerator(rootDir string) Generator {
	outputDir := filepath.Join(rootDir, "int/py")
	if err := resetDir(outputDir); err != nil {
		log.Fatal(err)
	}

	return &pythonGenerator{outputDir: outputDir}
}

//go:embed templates/py_*.tmpl
var pyTemplates embed.FS

// TODO(ENG-416): Specify variables, if defined (e.g. Redis nil).
// TODO(ENG-417): Specify types of input arguments per function.
// TODO(ENG-418): Function docstring instead of pass, if details are available.
// TODO(ENG-419): Specify output types, and define classes when needed.
func (g pythonGenerator) Generate(akURL string, n int, i *integrationsv1.Integration) {
	funcMap := template.FuncMap{
		"trimSpace": strings.TrimSpace,
		"refLabel":  stripNumPrefix,
	}
	t, err := template.New("").Funcs(funcMap).ParseFS(pyTemplates, "templates/py_*.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	b := new(bytes.Buffer)
	if err := t.ExecuteTemplate(b, "py_module.tmpl", i); err != nil {
		log.Fatal(err)
	}

	path := filepath.Join(g.outputDir, i.UniqueName+".py")
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
}
