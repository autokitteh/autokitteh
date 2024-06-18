package flowchartrt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/txtar"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var srcfs = kittehs.Must1(kittehs.TxtarToFS(txtar.Parse([]byte(`meow, world!
bub and bob say bye bye!

-- main.flowchart.yaml --
version: v1

values:
  cat: "meow"
  dog: "woof"

imports:
  - path: sub.flowchart.yaml
    name: sub

nodes:
  - name: n1    
    title: "main node"
    action:
      print:
          value: "meow, world!"
    goto: n2
  - name: n2
    title: "end node"
    action:
      print:
          value: "bub and bob say bye bye!"

-- sub.flowchart.yaml --
version: v1

values:
  woof: "woof"
  meow: "meow"

nodes:
  - name: s1
    title: "sub node"
    action:
      print:
        value: "woof"
    goto: s2
  - name: s2
    title: "sub end node"
`))))

func testBuild() (map[string][]byte, error) {
	a, err := rt{}.Build(
		context.Background(),
		srcfs,
		"main.flowchart.yaml",
		nil,
	)
	if err != nil {
		return nil, err
	}

	return a.CompiledData(), nil
}

func TestBuild(t *testing.T) {
	compiled, err := testBuild()
	require.NoError(t, err)

	// TODO
	assert.NotNil(t, compiled["main.flowchart.yaml"])
	assert.NotNil(t, compiled["sub.flowchart.yaml"])
}
