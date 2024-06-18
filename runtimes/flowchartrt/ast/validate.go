package ast

import (
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type validater interface {
	Validate() error
	name() sdktypes.Symbol
	kind() string
}

func validate(x validater) error {
	if err := x.Validate(); err != nil {
		name := x.name().String()
		if name != "" {
			name += ": "
		}
		return fmt.Errorf("%s%w", name, err)
	}

	return nil
}

func validateAll[T validater](xs []T) error {
	var errs []error

	for i, x := range xs {
		if err := validate(x); err != nil {
			errs = append(errs, fmt.Errorf("%s #%d: %w", x.kind(), i, err))
		}
	}

	return errors.Join(errs...)
}

func (f *Flowchart) Validate() error {
	res, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(JSONSchemaString),
		gojsonschema.NewGoLoader(f),
	)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !res.Valid() {
		return errors.Join(kittehs.Transform(res.Errors(), func(err gojsonschema.ResultError) error {
			return fmt.Errorf("%s: %s", err.Field(), err.Description())
		})...)
	}

	var errs []error

	if f.Version != Version {
		errs = append(errs, fmt.Errorf("unsupported version: %q", f.Version))
	}

	nodes := make(map[string]bool)
	for _, n := range f.Nodes {
		if nodes[n.Name] {
			errs = append(errs, fmt.Errorf("duplicate node: %q", n.Name))
			continue
		}

		nodes[n.Name] = true
	}

	errs = append(errs, validateAll(f.Imports))
	errs = append(errs, validateAll(f.Nodes))

	return errors.Join(errs...)
}

func (l *Import) name() sdktypes.Symbol { return kittehs.Must1(sdktypes.StrictParseSymbol(l.Name)) }
func (l *Import) kind() string          { return "load" }

func (l *Import) Validate() error {
	var errs []error

	if l.Name == "" {
		errs = append(errs, errors.New("missing name"))
	}

	if l.Path == "" {
		errs = append(errs, errors.New("missing path"))
	}

	return errors.Join(errs...)
}

func (n *Node) name() sdktypes.Symbol { return kittehs.Must1(sdktypes.StrictParseSymbol(n.Name)) }
func (n *Node) kind() string          { return "node" }

func (n *Node) Validate() error {
	var errs []error

	if n.Name == "" {
		errs = append(errs, errors.New("missing name"))
	}

	if n.Action != nil {
		if err := validate(n.Action); err != nil {
			errs = append(errs, fmt.Errorf("action: %w", err))
		}
	}

	if err := validateExpr(n.Goto); err != nil {
		errs = append(errs, fmt.Errorf("goto: %w", err))
	}

	return errors.Join(errs...)
}

func validateExpr(any) error {
	// TODO
	return nil
}

func (a *Action) name() sdktypes.Symbol { return sdktypes.InvalidSymbol }
func (a *Action) kind() string          { return "action" }

func (a *Action) Validate() error {
	if a == nil {
		return nil
	}

	switch {
	case a.Print != nil:
		return validate(a.Print)
	case a.Call != nil:
		return validate(a.Call)
	case a.Switch != nil:
		return validate(a.Switch)
	case a.ForEach != nil:
		return validate(a.ForEach)
	default:
		return errors.New("missing action")
	}
}

func (a *PrintAction) name() sdktypes.Symbol { return sdktypes.InvalidSymbol }
func (a *PrintAction) kind() string          { return "print action" }

func (a *PrintAction) Validate() error {
	var errs []error

	if a.Value == "" {
		errs = append(errs, errors.New("missing value"))
	} else if err := validateExpr(a.Value); err != nil {
		errs = append(errs, fmt.Errorf("value: %w", err))
	}

	return errors.Join(errs...)
}

func (a *CallAction) name() sdktypes.Symbol { return sdktypes.InvalidSymbol }
func (a *CallAction) kind() string          { return "call action" }

func (a *CallAction) Validate() error {
	var errs []error

	if a.Target == "" {
		errs = append(errs, errors.New("missing target"))
	}

	return errors.Join(errs...)
}

func (a SwitchAction) name() sdktypes.Symbol { return sdktypes.InvalidSymbol }
func (a SwitchAction) kind() string          { return "switch action" }

func (a SwitchAction) Validate() error {
	var errs []error

	errs = append(errs, validateAll(a))

	return errors.Join(errs...)
}

func (a *SwitchActionCase) name() sdktypes.Symbol { return sdktypes.InvalidSymbol }
func (a *SwitchActionCase) kind() string          { return "switch action case" }

func (a *SwitchActionCase) Validate() error {
	var errs []error

	if a.If != "" {
		if err := validateExpr(a.If); err != nil {
			errs = append(errs, fmt.Errorf("condition: %w", err))
		}
	}

	if a.Goto == "" {
		return errors.New("missing goto")
	}

	if err := validateExpr(a.Goto); err != nil {
		errs = append(errs, fmt.Errorf("goto: %w", err))
	}

	return errors.Join(errs...)
}

func (a *ForEachAction) name() sdktypes.Symbol { return sdktypes.InvalidSymbol }
func (a *ForEachAction) kind() string          { return "loop action" }

func (a *ForEachAction) Validate() error {
	var errs []error

	if a.Call == nil {
		errs = append(errs, errors.New("missing call"))
	} else if err := a.Call.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("call: %w", err))
	}

	if a.Items == "" {
		errs = append(errs, errors.New("missing items"))
	} else {
		if err := validateExpr(a.Items); err != nil {
			errs = append(errs, errors.New("missing items"))
			errs = append(errs, fmt.Errorf("over: %w", err))
		}
	}

	return errors.Join(errs...)
}
