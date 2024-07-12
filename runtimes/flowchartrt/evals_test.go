package flowchartrt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestEvalExprs(t *testing.T) {
	compiled, err := testBuild()
	require.NoError(t, err)

	r, err := testRun(compiled)
	require.NoError(t, err)

	args := map[string]sdktypes.Value{
		"one": sdktypes.NewIntegerValue(1),
		"two": sdktypes.NewIntegerValue(2),
		"st": kittehs.Must1(sdktypes.NewStructValue(sdktypes.Nothing, map[string]sdktypes.Value{
			"one": sdktypes.NewIntegerValue(1),
			"two": sdktypes.NewIntegerValue(2),
		})),
		"d": kittehs.Must1(sdktypes.NewStructValue(sdktypes.Nothing, map[string]sdktypes.Value{"one": sdktypes.NewIntegerValue(1)})),
	}

	th := kittehs.Must1(r.(*run).newThread(r.Values()["n1"], args))

	tests := []struct {
		expr string
		want sdktypes.Value
	}{
		{
			expr: "1 + 1",
			want: sdktypes.NewIntegerValue(2),
		},
		{
			expr: "1 == 1",
			want: sdktypes.True,
		},
		{
			expr: "false",
			want: sdktypes.False,
		},
		{
			expr: "args.one + args.one",
			want: sdktypes.NewIntegerValue(2),
		},
		{
			expr: "args.st.one + 1",
			want: sdktypes.NewIntegerValue(2),
		},
		{
			expr: "args.st.one + 1",
			want: sdktypes.NewIntegerValue(2),
		},
		{
			expr: `args.st.two + args.d["one"]`,
			want: sdktypes.NewIntegerValue(3),
		},
		{
			expr: "imports.sub.s1",
			want: r.(*run).modules["sub.flowchart.yaml"].exports["s1"],
		},
		{
			expr: "imports.sub.meow",
			want: sdktypes.NewStringValue("meow"),
		},
		{
			expr: "imports.sub.meow + imports.sub.woof",
			want: sdktypes.NewStringValue("meowwoof"),
		},
		{
			expr: "values.cat == imports.sub.meow",
			want: sdktypes.True,
		},
		{
			expr: "nodes.n1",
			want: r.Values()["n1"],
		},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			v, err := th.evalCELExpr(context.TODO(), test.expr, false, nil)
			if assert.NoError(t, err) {
				assert.True(t, test.want.Equal(v), v)
			}
		})
	}
}
