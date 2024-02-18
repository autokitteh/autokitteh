package apply

import (
	"encoding/json"

	"github.com/invopop/jsonschema"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var (
	JSONSchema       = jsonschema.Reflect(Root{})
	JSONSchemaString = string(kittehs.Must1(json.MarshalIndent(JSONSchema, "", "  ")))
)

const version = "v1"

type namer interface {
	comparable // allow to compare to empty.

	name() string
}

type Root struct {
	Version string `yaml:"version" json:"version" jsonschema:"required"`

	Connection *Connection `yaml:"connection,omitempty" json:"connection,omitempty"`
	Env        *Env        `yaml:"env,omitempty" json:"env,omitempty"`
	Project    *Project    `yaml:"project,omitempty" json:"project,omitempty"`

	Connections []*Connection `yaml:"connections,omitempty" json:"connections,omitempty"`
	Envs        []*Env        `yaml:"envs,omitempty" json:"envs,omitempty"`
	Projects    []*Project    `yaml:"projects,omitempty" json:"projects,omitempty"`
}

type User struct {
	Name string `yaml:"name" json:"name" jsonschema:"required"`
}

func (x *User) name() string {
	return x.Name
}

type Org struct {
	// TODO: Owner   string   `yaml:"owner"`
	Name    string   `yaml:"name" json:"name" jsonschema:"required"`
	Members []string `yaml:"members,omitempty" json:"members,omitempty"`
}

func (x *Org) name() string { return x.Name }

type Connection struct {
	Project     string `yaml:"project,omitempty" json:"project,omitempty"` // must be empty if under project.
	Name        string `yaml:"name" json:"name" jsonschema:"required"`
	Integration string `yaml:"integration" json:"integration" jsonschema:"required"`
	Token       string `yaml:"token,omitempty" json:"token,omitempty"` // if TokenEnvVar is set, use this as default.
	TokenEnvVar string `yaml:"token_env_var,omitempty" json:"token_env_var,omitempty"`
}

func (x *Connection) name() string { return x.Name }

type Project struct {
	Name        string        `yaml:"name" json:"name" jsonschema:"required"`
	Paths       []string      `yaml:"paths,omitempty" json:"paths"`
	Connections []*Connection `yaml:"connections,omitempty" json:"connections,omitempty"`
	Envs        []*Env        `yaml:"envs,omitempty" json:"envs,omitempty"`
}

func (x *Project) name() string { return x.Name }

type Env struct {
	Project  string     `yaml:"project,omitempty" json:"project,omitempty"` // must be empty if under project.
	Name     string     `yaml:"name" json:"name" jsonschema:"required"`
	Vars     []*Var     `yaml:"vars,omitempty" json:"vars,omitempty"`
	Mappings []*Mapping `yaml:"mappings,omitempty" json:"mappings,omitempty"`
}

func (x *Env) name() string { return x.Name }

type Var struct {
	Name     string `yaml:"name" json:"name" jsonschema:"required"`
	Value    string `yaml:"value,omitempty" json:"value,omitempty"` // if EnvVar is set, used as default value if env not found.
	IsSecret bool   `yaml:"is_secret,omitempty" json:"is_secret,omitempty"`
	EnvVar   string `yaml:"env_var,omitempty" json:"env_var,omitempty"` // if set, value is fetched from env.
}

func (x *Var) name() string { return x.Name }

type Mapping struct {
	Name       string          `yaml:"name" json:"name" jsonschema:"required"`
	Connection string          `yaml:"connection" json:"connection" jsonschema:"required"`
	Events     []*MappingEvent `yaml:"events,omitempty" json:"events,omitempty"`
}

func (x *Mapping) name() string { return x.Name }

type MappingEvent struct {
	EventType  string `yaml:"type" json:"type" jsonschema:"required"`
	EntryPoint string `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`
}
