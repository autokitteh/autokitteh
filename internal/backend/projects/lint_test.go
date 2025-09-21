package projects

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/internal/backend/projectsgrpcsvc"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func initialManifest() *manifest.Manifest {
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
					Webhook:   &struct{}{},
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

func createResources(fileName, funcName string) map[string][]byte {
	codeTmpl := `
def %s(event):
	pass
	`
	code := fmt.Appendf(nil, codeTmpl, funcName)
	m := map[string][]byte{
		fileName: code,
	}

	return m
}

func Test_checkHandlers(t *testing.T) {
	m := initialManifest()

	fileName, funcName, found := strings.Cut(m.Project.Triggers[0].Call, ":")
	require.Truef(t, found, "bad call - %q", m.Project.Triggers[0].Call)
	resources := createResources(fileName, funcName)
	vs := checkHandlers(sdktypes.InvalidProjectID, m, resources)
	require.Equal(t, 0, len(vs))

	m.Project.Triggers[0].Filter = "hiss"
	vs = checkHandlers(sdktypes.InvalidProjectID, m, resources)
	require.Equal(t, 1, len(vs))
	require.Equal(t, sdktypes.InvalidEventFilterRuleID, vs[0].RuleId)

	m.Project.Triggers[0].Filter = ""
	resources = createResources(fileName, funcName+"ZZZ")
	vs = checkHandlers(sdktypes.InvalidProjectID, m, resources)
	require.Equal(t, 1, len(vs))
	require.Equal(t, sdktypes.MissingHandlerRuleID, vs[0].RuleId)

	m.Project.Triggers[0].Call = "README.md:meow"

	vs = checkHandlers(sdktypes.InvalidProjectID, m, map[string][]byte{"README.md": []byte("meow")})
	require.Equal(t, 1, len(vs))
	require.Equal(t, sdktypes.FileCannotExportRuleID, vs[0].RuleId)
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

func TestNilSafe(t *testing.T) {
	Validate(sdktypes.InvalidProjectID, []byte("version: 1"), nil) // should not panic
}

func TestPyRequirements(t *testing.T) {
	test := func(lines ...string) []*sdktypes.CheckViolation {
		rs := map[string][]byte{"requirements.txt": []byte(strings.Join(lines, "\n"))}
		return checkPyRequirements(sdktypes.InvalidProjectID, nil, rs)
	}

	assert.Len(t, test("flask", "numpy==1.23.4", "# meow"), 0)
	assert.Len(t, test("1234"), 1)
	assert.Len(t, test("!meow"), 1)
	assert.Len(t, test("requests"), 1)
	assert.Len(t, test("requests==1"), 1)
}
