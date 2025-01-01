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
	Name        string        `yaml:"name,omitempty" json:"name,omitempty"`
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
	ParentKey string `yaml:"-" json:"-"` // associated with project or connection.

	Name   string `yaml:"name" json:"name" jsonschema:"required"`
	Value  string `yaml:"value" json:"value"`
	Secret bool   `yaml:"secret,omitempty" json:"secret,omitempty"`
}

func (v Var) GetKey() string { return v.ParentKey + "/" + v.Name }

type Trigger struct {
	ProjectKey string `yaml:"-" json:"-"` // associated with project.

	Name      string `yaml:"name" json:"name"`
	EventType string `yaml:"event_type,omitempty" json:"event_type,omitempty"`
	Filter    string `yaml:"filter,omitempty" json:"filter,omitempty"`

	Type          string    `yaml:"type,omitempty" json:"type,omitempty" jsonschema:"enum=schedule,enum=webhook,enum=connection"`
	Schedule      *string   `yaml:"schedule,omitempty" json:"schedule,omitempty"`
	Webhook       *struct{} `yaml:"webhook,omitempty" json:"webhook,omitempty"`
	ConnectionKey *string   `yaml:"connection,omitempty" json:"connection,omitempty" `

	Call string `yaml:"call,omitempty" json:"call,omitempty"`
}

func (t Trigger) GetKey() string {
	var id string

	if t.ProjectKey != "" {
		id = t.ProjectKey + ":"
	}

	what := ""

	if t.Schedule != nil {
		what = "schedule:" + *t.Schedule
	} else if t.Webhook != nil {
		what = "webhook"
	} else if t.ConnectionKey != nil {
		what = "connection:" + *t.ConnectionKey
	}

	return id + what + "/" + t.Name
}
