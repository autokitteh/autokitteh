// AK is the autokitteh command-line interface and local server.
package main

import (
	"os"
	"runtime"
	"runtime/pprof"

	"go.autokitteh.dev/autokitteh/cmd/ak/cmd"
)

func main() {
	f, err := os.Create("pprof2.cpu")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	m, err := os.Create("pprof2.mem")
	if err != nil {
		panic(err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(m); err != nil {
		panic(err)
	}

	cmd.Execute()
}
