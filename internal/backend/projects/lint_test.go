package projects

import (
	"fmt"
	"strings"
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
			Connections: []*manifest.Connection{
				{
					Name: "conn1",
					Vars: []*manifest.Var{
						{Name: "TOK", Value: "s3cr3t"},
					},
				},
			},
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

func Test_checkTriggerNames(t *testing.T) {
	m := initialManifest()
	m.Project.Triggers = []*manifest.Trigger{
		{Name: "A"},
		{Name: "B"},
	}
	vs := checkTriggerNames(sdktypes.InvalidProjectID, m, nil)
	require.Len(t, vs, 0)

	m.Project.Triggers = []*manifest.Trigger{
		{Name: "A"},
		{Name: "B"},
		{Name: "B"},
		{Name: "C"},
		{Name: "A"},
	}
	vs = checkTriggerNames(sdktypes.InvalidProjectID, m, nil)
	require.Len(t, vs, 2)
}

func createResources(fileName, fnName string) map[string][]byte {
	codeTmpl := `
def %s(event):
	pass
	`
	code := []byte(fmt.Sprintf(codeTmpl, fnName))
	m := map[string][]byte{
		fileName: code,
	}

	return m
}

func Test_checkHandlers(t *testing.T) {
	m := initialManifest()

	fileName, fnName, found := strings.Cut(m.Project.Triggers[0].Call, ":")
	require.Truef(t, found, "bad call - %q", m.Project.Triggers[0].Call)
	resources := createResources(fileName, fnName)
	vs := checkHandlers(sdktypes.InvalidProjectID, m, resources)
	require.Equal(t, 0, len(vs))

	resources = createResources(fileName, fnName+"ZZZ")
	vs = checkHandlers(sdktypes.InvalidProjectID, m, resources)
	require.Equal(t, 1, len(vs))
}

func Test_pyExports(t *testing.T) {
	code := []byte(`
def fn():
	pass

class cls():
	pass

# def x
define = 3
`)

	names, err := pyExports(code)
	require.NoError(t, err)
	require.Equal(t, []string{"fn", "cls"}, names)
}

func Test_pyConnCalls(t *testing.T) {
	code := []byte(`
client = boto3_client("aws1")
print("hello")
conn = asana_client('asana2')
g = gmail_client(conn1)
`)

	conns := pyConnCalls(code)
	expected := []connCall{
		{"aws1", 2},
		{"asana2", 4},
		{"conn1", 5},
	}
	require.Equal(t, expected, conns)
}

func Test_checkCodeConnections(t *testing.T) {
	m := initialManifest()
	codeTmpl := `
	boto3_client("%s")
	`

	code := []byte(fmt.Sprintf(codeTmpl, m.Project.Connections[0].Name))
	fileName := "handler.py"
	resources := map[string][]byte{
		fileName: code,
	}
	vs := checkCodeConnections(sdktypes.InvalidProjectID, m, resources)
	require.Len(t, vs, 0)

	code = []byte(fmt.Sprintf(codeTmpl, m.Project.Connections[0].Name+"ZZZ"))
	resources[fileName] = code
	vs = checkCodeConnections(sdktypes.InvalidProjectID, m, resources)
	require.Len(t, vs, 1)
}
