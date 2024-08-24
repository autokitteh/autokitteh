// Description of the manifest schema. For detailed field explanations, see /docs/autokitteh.yaml.
package manifest

import (
	"encoding/json"

	"github.com/invopop/jsonschema"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest/internal/actions"
)

type (
	Action  = actions.Action
	Actions = actions.Actions
)

var (
	JSONSchema       = jsonschema.Reflect(Manifest{})
	JSONSchemaString = string(kittehs.Must1(json.MarshalIndent(JSONSchema, "", "  ")))
)

const Version = "v1"

type stringKeyer string

func (s stringKeyer) GetKey() string { return string(s) }

type Manifest struct {
	Version string   `yaml:"version,omitempty" json:"version,omitempty" jsonschema:"required"`
	Project *Project `yaml:"project,omitempty" json:"project,omitempty"`
}

type Project struct {
	Name        string        `yaml:"name" json:"name"`
	Connections []*Connection `yaml:"connections,omitempty" json:"connections,omitempty"`
	Triggers    []*Trigger    `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Vars        []*Var        `yaml:"vars,omitempty" json:"vars,omitempty"`
}

func (p Project) GetKey() string { return p.Name }

type Connection struct {
	ProjectKey string `yaml:"-" json:"-"` // belongs to project.

	Name           string `yaml:"name" json:"name" jsonschema:"required"`
	IntegrationKey string `yaml:"integration" json:"integration" jsonschema:"required"`
	Vars           []*Var `yaml:"vars,omitempty" json:"vars,omitempty"`
}

func (c Connection) GetKey() string { return c.ProjectKey + "/" + c.Name }

type Var struct {
	ParentKey string `yaml:"-" json:"-"` // associated with env or connection.

	Name        string `yaml:"name" json:"name" jsonschema:"required"`
	Value       string `yaml:"value,omitempty" json:"value,omitempty"`
	Secret      bool   `yaml:"secret,omitempty" json:"secret,omitempty"`
	Optional    bool   `yaml:"optional,omitempty" json:"optional,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

func (v Var) GetKey() string { return v.ParentKey + "/" + v.Name }

type Trigger struct {
	EnvKey string `yaml:"-" json:"-"` // associated with env.

	ConnectionKey string `yaml:"connection,omitempty" json:"connection,omitempty"` // jsonscheme: FIXME: ENG-862
	Name          string `yaml:"name" json:"name"`
	EventType     string `yaml:"event_type,omitempty" json:"event_type,omitempty"`
	Filter        string `yaml:"filter,omitempty" json:"filter,omitempty"`

	// for scheduled trigger. Schedule could be passed in `data` section as well
	Schedule string `yaml:"schedule,omitempty" json:"schedule,omitempty"` // jsonscheme: FIXME: ENG-862

	// Arbitrary data to be passed with the trigger.
	// The dispatcher can use this data, for example, to extract HTTP path parameters.
	// For example: `data: { "path": "/a/{b}/{c...}"}`, if the connection is an HTTP connection.
	Data map[string]any `yaml:"data,omitempty" json:"data,omitempty"`

	Call       string `yaml:"call,omitempty" json:"call,omitempty" jsonschema:"oneof_required=call"`
	Entrypoint string `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty" jsonschema:"oneof_required=entrypoint"`
}

func (t Trigger) GetKey() string {
	var id string

	if t.EnvKey != "" {
		id = t.EnvKey + ":"
	}

	return id + t.ConnectionKey + "/" + t.Name
}
