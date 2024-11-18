package projects

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.autokitteh.dev/autokitteh/internal/backend/projectsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func initialManifest() *manifest.Manifest {
	var wh struct{}
	return &manifest.Manifest{
		Version: "v1",
		Project: &manifest.Project{
			Name: "test",
			Triggers: []*manifest.Trigger{
				{
					Name:      "events",
					EventType: "post",
					Call:      "program.py:on_event",
					Webhook:   &wh,
				},
			},
			Vars: []*manifest.Var{
				{Name: "USER", Value: "garfield"},
			},
		},
	}
}

func Test_checkNoTriggers(t *testing.T) {
	m := initialManifest()

	vs := checkNoTriggers(sdktypes.InvalidProjectID, m, nil)
	require.Len(t, vs, 0)
	m.Project.Triggers = nil

	vs = checkNoTriggers(sdktypes.InvalidProjectID, m, nil)
	require.Len(t, vs, 1)
}

func Test_checkEmptyVars(t *testing.T) {
	m := initialManifest()

	vs := checkEmptyVars(sdktypes.InvalidProjectID, m, nil)
	require.Len(t, vs, 0)

	m.Project.Vars[0].Value = ""
	vs = checkEmptyVars(sdktypes.InvalidProjectID, m, nil)
	require.Len(t, vs, 1)

	m.Project.Vars = append(m.Project.Vars, &manifest.Var{Name: "TOKEN", Value: ""})
	m.Project.Vars = append(m.Project.Vars, &manifest.Var{Name: "HOME", Value: "/home/ak"})
	vs = checkEmptyVars(sdktypes.InvalidProjectID, m, nil)
	require.Len(t, vs, 2)
}

func Test_checkSize(t *testing.T) {
	maxProjectSize := projectsgrpcsvc.Configs.Default.MaxUploadSize

	resources := make(map[string][]byte)
	resources["one"] = make([]byte, 100)
	resources["two"] = make([]byte, 200)

	m := initialManifest()
	vs := checkSize(sdktypes.InvalidProjectID, m, resources)
	require.Len(t, vs, 0)

	resources["three"] = make([]byte, maxProjectSize)
	vs = checkSize(sdktypes.InvalidProjectID, m, resources)
	require.Len(t, vs, 1)
}

func Test_checkConnectionNames(t *testing.T) {
	m := initialManifest()
	m.Project.Connections = []*manifest.Connection{
		{Name: "A"},
		{Name: "B"},
	}
	vs := checkConnectionNames(sdktypes.InvalidProjectID, m, nil)
	require.Len(t, vs, 0)

	m.Project.Connections = []*manifest.Connection{
		{Name: "A"},
		{Name: "B"},
		{Name: "B"},
		{Name: "C"},
		{Name: "A"},
	}
	vs = checkConnectionNames(sdktypes.InvalidProjectID, m, nil)
	require.Len(t, vs, 2)
}
