// AK is the autokitteh command-line interface and local server.
package main

import (
	"os"

	"go.autokitteh.dev/autokitteh/cmd/ak/cmd"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

func main() {
	common.SetWriters(os.Stdout, os.Stderr)
	cmd.Execute()
}
