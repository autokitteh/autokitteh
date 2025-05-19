//go:build enterprise
// +build enterprise

package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

type model struct {
	ImportPath string
	PkgName    string
	Name       string
}

func main() {
	stmts, err := gormschema.New("postgres").Load(&scheme.WorkerInfo{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
