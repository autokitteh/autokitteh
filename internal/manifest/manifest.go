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
	Version string `yaml:"version,omitempty" json:"version,omitempty" jsonschema:"required"`

	Project *Project `yaml:"project,omitempty" json:"project,omitempty"`
}

type Project struct {
	Name        string        `yaml:"name" json:"name" jsonschema:"required"`
	Connections []*Connection `yaml:"connections,omitempty" json:"connections,omitempty"`
	Triggers    []*Trigger    `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Vars        []*EnvVar     `yaml:"vars,omitempty" json:"vars,omitempty"`
}

func (p Project) GetKey() string { return p.Name }

type Connection struct {
	ProjectKey string `yaml:"-" json:"-"` // belongs to project.

	Name           string `yaml:"name" json:"name" jsonschema:"required"`
	Token          string `yaml:"token" json:"token" jsonschema:"required"`
	IntegrationKey string `yaml:"integration" json:"integration" jsonschema:"required"`
}

func (c Connection) GetKey() string { return c.ProjectKey + "/" + c.Name }

type EnvVar struct {
	EnvKey string `yaml:"-" json:"-"` // associated with env.

	Name     string `yaml:"name" json:"name" jsonschema:"required"`
	Value    string `yaml:"value,omitempty" json:"value,omitempty"` // if EnvVar is set, used as default value if env not found.
	IsSecret bool   `yaml:"is_secret,omitempty" json:"is_secret,omitempty"`
	EnvVar   string `yaml:"env_var,omitempty" json:"env_var,omitempty"` // if set, value is fetched from env.
}

func (v EnvVar) GetKey() string { return v.EnvKey + "/" + v.Name }

type Trigger struct {
	EnvKey string `yaml:"-" json:"-"` // associated with env.

	ConnectionKey string `yaml:"connection" json:"connection" jsonschema:"required"` // comming from connection.
	EventType     string `yaml:"event_type" json:"event_type" jsonschema:"required"`
	Entrypoint    string `yaml:"entrypoint" json:"entrypoint" jsonschema:"required"`
}

func (t Trigger) GetKey() string {
	id := t.ConnectionKey + "/" + t.EventType
	if t.EnvKey != "" {
		id = t.EnvKey + ":" + id
	}

	return id
}
