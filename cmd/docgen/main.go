// DocGen is an internal tool that auto-generates Docusaurus markdown
// files about the AutoKitteh CLI tool for https://docs.autokitteh.com.
package main

import (
	"fmt"
	"os"
)

const (
	outputDir = "gen"
)

func main() {
	if err := resetDir(outputDir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// TODO(ENG-415): Generate MD or MDX files for each CLI command.
	// See: https://github.com/spf13/cobra/tree/main/doc
}

func resetDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755) // rwxr-xr-x
}
