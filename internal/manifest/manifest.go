package manifest

import (
	"crypto/md5"
	"encoding/hex"
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
	Name string `yaml:"name" json:"name" jsonschema:"required"`

	Vars []*EnvVar `yaml:"vars,omitempty" json:"vars,omitempty"`

	Connections []*Connection `yaml:"connections,omitempty" json:"connections,omitempty"`

	Triggers []*Trigger `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	HTTP     *HTTP      `yaml:"http,omitempty" json:"http,omitempty"`
}

func (p Project) GetKey() string { return p.Name }

type Connection struct {
	ProjectKey string `yaml:"-" json:"-"` // belongs to project.

	Name           string `yaml:"name" json:"name" jsonschema:"required"`
	Token          string `yaml:"token" json:"token"`
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

	ConnectionKey string         `yaml:"connection" json:"connection" jsonschema:"required"` // coming from connection.
	Name          string         `yaml:"name,omitempty" json:"name,omitempty"`
	EventType     string         `yaml:"event_type,omitempty" json:"event_type,omitempty"`
	Filter        string         `yaml:"filter,omitempty" json:"filter,omitempty"`
	Data          map[string]any `yaml:"data,omitempty" json:"additional_data,omitempty"`
	Type          string         `yaml:"type" json:"type,omitempty"`

	Call       string `yaml:"call,omitempty" json:"call,omitempty" jsonschema:"oneof_required=call"`
	Entrypoint string `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty" jsonschema:"oneof_required=entrypoint"`
}

func (t Trigger) GetKey() string {
	var id string
	if t.EnvKey != "" {
		id = t.EnvKey + ":"
	}

	id += t.ConnectionKey + "/"

	if t.Name == "" {
		hash := md5.Sum(kittehs.Must1(json.Marshal(t)))
		return id + hex.EncodeToString(hash[:])
	}

	return id + t.Name
}

type HTTP struct {
	Disabled bool         `yaml:"disable,omitempty" json:"disable,omitempty"`
	Routes   []*HTTPRoute `yaml:"routes,omitempty" json:"routes,omitempty"`
}

type HTTPRoute struct {
	Name   string `yaml:"name,omitempty" json:"name,omitempty"`
	Path   string `yaml:"path,omitempty" json:"path,omitempty"`
	Method string `yaml:"method,omitempty" json:"method,omitempty"`

	Call       string `yaml:"call,omitempty" json:"call,omitempty" jsonschema:"oneof_required=call"`
	Entrypoint string `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty" jsonschema:"oneof_required=entrypoint"`
}
