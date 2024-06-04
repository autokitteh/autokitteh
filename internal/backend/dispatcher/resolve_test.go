package dispatcher

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbfactory"
	"go.autokitteh.dev/autokitteh/internal/backend/envs"
	"go.autokitteh.dev/autokitteh/internal/backend/projects"
	wf "go.autokitteh.dev/autokitteh/internal/backend/workflows"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestResolveEnv(t *testing.T) {
	S := kittehs.Must11(sdktypes.ParseSymbol)

	var pids [4]sdktypes.ProjectID
	for i := range pids {
		pids[i] = sdktypes.NewProjectID()
	}

	var eids [6]sdktypes.EnvID
	for i := range eids {
		eids[i] = sdktypes.NewEnvID()
	}

	objs := []sdktypes.Object{
		sdktypes.NewProject().WithID(pids[0]).WithName(S("p0")),
		sdktypes.NewEnv().WithID(eids[0]).WithName(S("default")).WithProjectID(pids[0]),

		sdktypes.NewProject().WithID(pids[1]).WithName(S("p1")),
		sdktypes.NewEnv().WithID(eids[1]).WithName(S("default")).WithProjectID(pids[1]),
		sdktypes.NewEnv().WithID(eids[2]).WithName(S("other")).WithProjectID(pids[1]),

		sdktypes.NewProject().WithID(pids[2]).WithName(S("p2")),
		sdktypes.NewEnv().WithID(eids[3]).WithName(S("other")).WithProjectID(pids[2]),
		sdktypes.NewEnv().WithID(eids[4]).WithName(S("another")).WithProjectID(pids[2]),

		sdktypes.NewProject().WithID(pids[3]).WithName(S("p3")),
		sdktypes.NewEnv().WithID(eids[5]).WithName(S("iamspecial")).WithProjectID(pids[3]),
	}

	testdb := dbfactory.NewTest(t, objs)

	z := zap.NewNop()

	svcs := &wf.Services{
		Envs:     envs.New(z, testdb),
		Projects: &projects.Projects{DB: testdb, Z: z},
	}

	tests := []struct {
		in  string
		out sdktypes.EnvID
		err bool
	}{
		{
			in:  "p0",
			out: eids[0],
		},
		{
			in:  "p0/default",
			out: eids[0],
		},
		{
			in:  "p0/hiss",
			err: true,
		},
		{
			in:  "p1",
			out: eids[1],
		},
		{
			in:  "p1/default",
			out: eids[1],
		},
		{
			in:  "p1/other",
			out: eids[2],
		},
		{
			in:  "p1/hiss",
			err: true,
		},
		{
			in:  "p2",
			err: true,
		},
		{
			in:  "p2/other",
			out: eids[3],
		},
		{
			in:  "p2/another",
			out: eids[4],
		},
		{
			in:  "p3",
			out: eids[5],
		},
		{
			in:  "p3/default",
			err: true,
		},
		{
			in:  "p3/iamspecial",
			out: eids[5],
		},
	}

	ctx := context.Background()

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			got, err := resolveEnv(ctx, svcs, test.in)
			if err != nil {
				if !test.err {
					t.Errorf("resolveEnv() unexpected error = %v", err)
				}
			} else if got != test.out {
				t.Errorf("resolveEnv() got = %v, want %v", got, test.out)
			}
		})
	}
}
