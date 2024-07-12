package ast

import (
	"encoding/json"
	"strings"

	"github.com/invopop/jsonschema"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const Version = "v1"

var (
	JSONSchema       = jsonschema.Reflect(Flowchart{})
	JSONSchemaString = string(kittehs.Must1(json.MarshalIndent(JSONSchema, "", "  ")))
)

type Options struct {
	InterpolateAllStrings bool   `yaml:"interpolate_all_strings,omitempty" json:"interpolate_all_strings,omitempty"`
	DefaultInterpolation  string `yaml:"default_interpolation,omitempty" json:"default_interpolation,omitempty"`
}

type Flowchart struct {
	Version string         `yaml:"version" json:"version" jsonschema:"required"`
	Values  map[string]any `yaml:"values,omitempty" json:"values,omitempty"` // const - cannot be an expression.
	Imports []*Import      `yaml:"imports,omitempty" json:"imports,omitempty"`
	Nodes   []*Node        `yaml:"nodes,omitempty" json:"nodes,omitempty"`
	Memo    any            `yaml:"memo,omitempty" json:"memo,omitempty"`
	Options *Options       `yaml:"options,omitempty" json:"options,omitempty"`

	path string
}

func (f *Flowchart) SafeOptions() *Options {
	if f.Options == nil {
		return &Options{}
	}

	return f.Options
}

func (f *Flowchart) GetNode(name string) *Node {
	_, node := kittehs.FindFirst(f.Nodes, func(n *Node) bool { return n.Name == name })
	return node
}

type Import struct {
	Path string `yaml:"path" json:"path" jsonschema:"required"`
	Name string `yaml:"name" json:"name" jsonschema:"required"`
}

type Node struct {
	Name   string  `yaml:"name" json:"name" jsonschema:"required"`
	Title  string  `yaml:"title,omitempty" json:"title,omitempty"`
	Action *Action `yaml:"action,omitempty" json:"action,omitempty"`
	Result any     `yaml:"result,omitempty" json:"result,omitempty"`
	Goto   string  `yaml:"goto,omitempty" json:"goto,omitempty"`
	Memo   any     `yaml:"memo,omitempty" json:"memo,omitempty"`

	loc sdktypes.CodeLocation
}

func (n *Node) Location() sdktypes.CodeLocation { return n.loc }

type Action struct {
	Print   *PrintAction   `yaml:"print,omitempty" json:"print,omitempty" jsonschema:"oneof_required=print"`
	Call    *CallAction    `yaml:"call,omitempty" json:"call,omitempty" jsonschema:"oneof_required=call"`
	Switch  SwitchAction   `yaml:"switch,omitempty" json:"switch,omitempty" jsonschema:"oneof_required=switch"`
	ForEach *ForEachAction `yaml:"foreach,omitempty" json:"foreach,omitempty" jsonschema:"oneof_required=foreach"`
}

type PrintAction struct {
	Value any `yaml:"value" json:"value" jsonschema:"required"`
}

type CallAction struct {
	Target string         `yaml:"target" json:"target" jsonschema:"required"`
	Args   map[string]any `yaml:"args,omitempty" json:"args,omitempty"`

	Async         bool   `yaml:"async,omitempty" json:"async,omitempty"`     // only for nodes.
	Timeout       any    `yaml:"timeout,omitempty" json:"timeout,omitempty"` // only for non-nodes.
	OnTimeoutGoto string `yaml:"on_timeout_goto,omitempty" json:"on_timeout_goto,omitempty"`
}

type SwitchAction []*SwitchActionCase

type SwitchActionCase struct {
	If   any    `yaml:"if" json:"if" jsonschema:"required"`
	Goto string `yaml:"goto" json:"goto" jsonschema:"required"`
}

type ForEachAction struct {
	Call  *CallAction `yaml:"call" json:"call" jsonschema:"required"`
	Items any         `yaml:"items" json:"items" jsonschema:"required"`
}

func (f *Flowchart) Exports() ([]sdktypes.BuildExport, error) {
	var xs []sdktypes.BuildExport

	for _, node := range f.Nodes {
		if strings.HasPrefix(node.Name, "_") {
			continue
		}

		sym, err := sdktypes.StrictParseSymbol(node.Name)
		if err != nil {
			return nil, err
		}

		xs = append(xs, sdktypes.NewBuildExport().WithSymbol(sym))
	}

	for k := range f.Values {
		if strings.HasPrefix(k, "_") {
			continue
		}

		sym, err := sdktypes.StrictParseSymbol(k)
		if err != nil {
			return nil, err
		}

		xs = append(xs, sdktypes.NewBuildExport().WithSymbol(sym))
	}

	return xs, nil
}
