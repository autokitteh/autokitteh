//go:build !enterprise
// +build !enterprise

/*
Package test runs end-to-end "black-box" system tests on for the
autokitteh CLI tool, functioning both as a server and as a client.

It can also control other tools, dependencies, and in-memory fixtures
(e.g. Temporal, databases, caches, and HTTP webhooks).

Test cases are defined as [txtar] files in the [testdata] directory
tree. Their structure and scripting language is defined [here].

Other than local and CI/CD testing, this may be used for benchmarking,
profiling, and load/stress testing.

[txtar]: https://pkg.go.dev/golang.org/x/tools/txtar
[testdata]: https://github.com/autokitteh/autokitteh/tree/main/systest/testdata
[here]: https://github.com/autokitteh/autokitteh/tree/main/systest/README.md
*/
package systest

import (
	"strings"
	"testing"
)

const testFilesFilter = ".txtar"

func testFilter(name string) bool {
	return strings.HasSuffix(name, testFilesFilter) && !strings.HasSuffix(name, ".ee.txtar")
}

func setupExternalResources(t *testing.T) map[string]any {
	return map[string]any{}
}
