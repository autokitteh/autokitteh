package akmod

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/autokitteh/autokitteh/internal/pkg/statestore"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/sdk/pluginimpl"

	L "github.com/autokitteh/autokitteh/pkg/l"
)

const prefixSep = "|"

type state struct {
	projectID  apiproject.ProjectID
	stateStore statestore.Store
	l          L.L
	prefix     string
}

func (s *state) asValue(funcToValue pluginimpl.FuncToValueFunc) *apivalues.Value {
	simpleFuncToValue := func(n string, f pluginimpl.SimplePluginMethodFunc) *apivalues.Value {
		return funcToValue(n, pluginimpl.SimplifyPluginMethodFunc(f))
	}

	return apivalues.MustNewValue(apivalues.StructValue{
		Ctor: apivalues.Symbol("state"),
		Fields: map[string]*apivalues.Value{
			"set":     simpleFuncToValue("set", s.set),
			"get":     simpleFuncToValue("get", s.get),
			"inc":     simpleFuncToValue("inc", s.inc),
			"dec":     simpleFuncToValue("dec", s.dec),
			"insert":  simpleFuncToValue("insert", s.insert),
			"take":    simpleFuncToValue("take", s.take),
			"index":   simpleFuncToValue("index", s.index),
			"length":  simpleFuncToValue("length", s.length),
			"keys":    simpleFuncToValue("keys", s.keys),
			"set_key": simpleFuncToValue("set_key", s.setKey),
			"get_key": simpleFuncToValue("get_key", s.getKey),
			"scoped":  funcToValue("scope", s.scoped),
		},
	})
}

func (s *state) name(n string) (string, error) {
	if strings.Contains(n, prefixSep) {
		return "", fmt.Errorf("invalid name")
	}

	if s.prefix == "" {
		return n, nil
	}

	return fmt.Sprintf("%s%s%s", s.prefix, prefixSep, n), nil
}

func (s *state) set(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var (
		name string
		v    *apivalues.Value
	)

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name, "value", &v); err != nil {
		return nil, err
	}

	if v.IsEphemeral() {
		return nil, fmt.Errorf("ephemeral values cannot be persisted")
	}

	var err error
	if name, err = s.name(name); err != nil {
		return nil, err
	}

	if err := s.stateStore.Set(ctx, s.projectID, name, v); err != nil {
		return nil, fmt.Errorf("set: %w", err)
	}

	return apivalues.None, nil
}

func (s *state) get(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var (
		name string
		fail = true
	)

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name, "fail?", &fail); err != nil {
		return nil, err
	}

	var err error
	if name, err = s.name(name); err != nil {
		return nil, err
	}

	v, _, err := s.stateStore.Get(ctx, s.projectID, name)
	if err != nil {
		if !fail && errors.Is(err, statestore.ErrNotFound) {
			return apivalues.None, nil
		}

		return nil, fmt.Errorf("get: %w", err)
	}

	return v, nil

}

func (s *state) insert(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var (
		name  string
		index int
		value *apivalues.Value
	)

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name, "value", &value, "index?", &index); err != nil {
		return nil, err
	}

	if err := s.stateStore.Insert(ctx, s.projectID, name, index, value); err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}

	return apivalues.None, nil
}

func (s *state) dec(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var (
		name   string
		amount int64 = 1
	)

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name, "amount?", &amount); err != nil {
		return nil, err
	}

	v, err := s.stateStore.Inc(ctx, s.projectID, name, -amount)
	if err != nil {
		return nil, fmt.Errorf("inc: %w", err)
	}

	return v, nil
}

func (s *state) inc(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var (
		name   string
		amount int64 = 1
	)

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name, "amount?", &amount); err != nil {
		return nil, err
	}

	v, err := s.stateStore.Inc(ctx, s.projectID, name, amount)
	if err != nil {
		return nil, fmt.Errorf("inc: %w", err)
	}

	return v, nil
}

func (s *state) take(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var (
		name  string
		index int
		count = 1
	)

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name, "index?", &index, "count?", &count); err != nil {
		return nil, err
	}

	v, err := s.stateStore.Take(ctx, s.projectID, name, index, count)
	if err != nil {
		return nil, fmt.Errorf("take: %w", err)
	}

	return v, nil
}

func (s *state) index(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var (
		name  string
		index int
	)

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name, "index", &index); err != nil {
		return nil, err
	}

	v, err := s.stateStore.Index(ctx, s.projectID, name, index)
	if err != nil {
		return nil, fmt.Errorf("take: %w", err)
	}

	return v, nil
}

func (s *state) length(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var name string

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name); err != nil {
		return nil, err
	}

	n, err := s.stateStore.Length(ctx, s.projectID, name)
	if err != nil {
		return nil, fmt.Errorf("take: %w", err)
	}

	return apivalues.Integer(int64(n)), nil
}

func (s *state) keys(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var name string

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name); err != nil {
		return nil, err
	}

	return s.stateStore.Keys(ctx, s.projectID, name)
}

func (s *state) getKey(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var (
		name string
		key  *apivalues.Value
		fail = true
	)

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name, "key", &key, "fail?", &fail); err != nil {
		return nil, err
	}

	v, err := s.stateStore.GetKey(ctx, s.projectID, name, key)
	if err != nil {
		if !fail && errors.Is(err, statestore.ErrNotFound) {
			return apivalues.None, nil
		}

		return nil, fmt.Errorf("get_key: %w", err)
	}

	return v, nil
}

func (s *state) setKey(
	ctx context.Context,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
) (*apivalues.Value, error) {
	var (
		name       string
		key, value *apivalues.Value
	)

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name, "key", &key, "value", &value); err != nil {
		return nil, err
	}

	if err := s.stateStore.SetKey(ctx, s.projectID, name, key, value); err != nil {
		return nil, fmt.Errorf("set_key: %w", err)
	}

	return apivalues.None, nil
}

func (s *state) scoped(
	ctx context.Context,
	_ string,
	args []*apivalues.Value,
	kwargs map[string]*apivalues.Value,
	funcToValue pluginimpl.FuncToValueFunc,
) (*apivalues.Value, error) {
	var name string

	if err := pluginimpl.UnpackArgs(args, kwargs, "name", &name); err != nil {
		return nil, err
	}

	s1 := *s
	if s1.prefix != "" {
		s1.prefix += prefixSep
	}

	s1.prefix += name

	return s1.asValue(funcToValue), nil
}
