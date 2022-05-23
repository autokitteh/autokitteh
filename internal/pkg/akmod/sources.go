package akmod

import (
	"context"
	"fmt"
	"strings"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apivalues"
	"go.autokitteh.dev/sdk/pluginimpl"

	"github.com/autokitteh/L"
)

type binding struct {
	bindingName string
	classifiers []string
	fn          apivalues.FunctionValue
	context     *apivalues.Value
}

// box so that subsources will modify the actual bindings.
type bindings struct{ bindings []*binding }

type sources struct {
	srcBindingName string
	event          *apievent.Event

	bindings *bindings
	l        L.L
	subs     map[string]string
}

func isBindingAllowed(subs map[string]string, candidate string) bool {
	return subs == nil || subs[strings.SplitN(candidate, ".", 2)[0]] != "" || subs[candidate] != ""
}

func (s *sources) bind_(nameParts []string, target, context *apivalues.Value) error {
	name := strings.Join(nameParts, ".")

	switch v := target.Get().(type) {
	case apivalues.FunctionValue:
		s.bindings.bindings = append(s.bindings.bindings, &binding{
			bindingName: nameParts[0],
			classifiers: nameParts[1:],
			context:     context,
			fn:          v,
		})
		s.l.Debug("source bound", "name", name, "target", target, "context", context)
		return nil
	case apivalues.DictValue:
		for _, kv := range v {
			if err := s.bind_(append(nameParts, kv.K.String()), kv.V, context); err != nil {
				return err
			}
		}
		return nil
	case apivalues.StructValue:
		for k, v := range v.Fields {
			if err := s.bind_(append(nameParts, k), v, context); err != nil {
				return err
			}
		}
		return nil
	case apivalues.ModuleValue:
		for k, v := range v.Members {
			if err := s.bind_(append(nameParts, k), v, context); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported handler value %q for %q", target, name)
	}
}

func (s *sources) sub(
	_ context.Context,
	_ string,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	funcToValue pluginimpl.FuncToValueFunc,
) (*apivalues.Value, error) {
	if s.subs != nil {
		// TODO
		return nil, fmt.Errorf("nested subs are not allowed")
	}

	subs := make(map[string]string, len(args)+len(kwargs))

	for _, a := range args {
		v, ok := a.Get().(apivalues.StringValue)
		if !ok {
			return nil, fmt.Errorf("all arguments must be of string value")
		}

		s := v.String()

		subs[s] = s
	}

	for k, v := range kwargs {
		v, ok := v.Get().(apivalues.StringValue)
		if !ok {
			return nil, fmt.Errorf("all arguments must be of string value")
		}

		subs[k] = v.String()
	}

	if s.subs != nil {
		// make sure this is a subset of parent subs.

		for k := range subs {
			if !isBindingAllowed(s.subs, k) {
				return nil, fmt.Errorf("binding %q is not allowed", k)
			}
		}
	}

	s1 := *s
	s1.subs = subs

	return s1.asValue(funcToValue), nil
}

func (s *sources) bind(
	_ context.Context,
	_ string,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	_ pluginimpl.FuncToValueFunc,
) (*apivalues.Value, error) {
	var (
		src             string
		target, context *apivalues.Value
	)

	if err := pluginimpl.UnpackArgs(
		args, kwargs,
		"source", &src,
		"target", &target,
		"context?", &context,
	); err != nil {
		return nil, err
	}

	s.l.Debug("binding source", "source", src, "target", target, "context", context)

	if !isBindingAllowed(s.subs, src) {
		return nil, fmt.Errorf("binding %q is not permitted", src)
	}

	if v, ok := s.subs[src]; ok {
		src = v
	} else if parts := strings.SplitN(src, ".", 2); len(parts) == 2 {
		if v, ok := s.subs[parts[0]]; ok {
			src = fmt.Sprintf("%s.%s", v, parts[1])
		}
	}

	if err := s.bind_(strings.Split(src, "."), target, context); err != nil {
		return nil, err
	}

	return apivalues.None, nil
}

func (s *sources) match(
	_ context.Context,
	_ string,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	_ pluginimpl.FuncToValueFunc,
) (*apivalues.Value, error) {
	bindingName := s.srcBindingName
	eventType := s.event.Type()

	if err := pluginimpl.UnpackArgs(
		args, kwargs,
		"source?", &bindingName,
		"event_type?", &eventType,
	); err != nil {
		return nil, err
	}

	lv := apivalues.ListValue{}

	if !isBindingAllowed(s.subs, fmt.Sprintf("%s.%s", bindingName, eventType)) {
		return nil, fmt.Errorf("binding \"%s.%s\" is not allowed", bindingName, eventType)
	}

	for _, b := range s.bindings.bindings {
		if b.bindingName != bindingName {
			continue
		}

		if len(b.classifiers) > 0 {
			if eventType != "" && eventType != strings.Join(b.classifiers, ".") {
				continue
			}
		}

		context := b.context
		if context == nil {
			context = apivalues.None
		}

		lv = append(
			lv,
			apivalues.MustNewValue(
				apivalues.StructValue{
					Ctor: apivalues.String("binding"),
					Fields: map[string]*apivalues.Value{
						"name":    apivalues.String(bindingName),
						"handler": apivalues.MustNewValue(b.fn),
						"context": context,
					},
				},
			),
		)
	}

	return apivalues.MustNewValue(lv), nil
}

func (s *sources) asStruct(funcToValue pluginimpl.FuncToValueFunc) apivalues.StructValue {
	return apivalues.StructValue{
		Ctor: apivalues.Symbol("sources"),
		Fields: map[string]*apivalues.Value{
			"bind": funcToValue("bind", s.bind, pluginimpl.WithFlags("allow_passing_call_values")),
			"sub":  funcToValue("sub", s.sub),
		},
	}
}

func (s *sources) asValueWithMatch(funcToValue pluginimpl.FuncToValueFunc) *apivalues.Value {
	st := s.asStruct(funcToValue)

	st.Fields["match"] = funcToValue("match", s.match, pluginimpl.WithFlags("allow_passing_call_values"))

	return apivalues.MustNewValue(st)
}

func (s *sources) asValue(funcToValue pluginimpl.FuncToValueFunc) *apivalues.Value {
	return apivalues.MustNewValue(s.asStruct(funcToValue))
}
