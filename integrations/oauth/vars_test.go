package oauth

import (
	"context"
	"errors"
	"maps"
	"slices"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type fakeVarsService struct {
	data map[sdktypes.VarScopeID]map[sdktypes.Symbol]sdktypes.Var
}

var _ sdkservices.Vars = &fakeVarsService{}

func newFakeVars() *fakeVarsService {
	return &fakeVarsService{
		data: make(map[sdktypes.VarScopeID]map[sdktypes.Symbol]sdktypes.Var),
	}
}

func (s *fakeVarsService) Set(ctx context.Context, vs ...sdktypes.Var) error {
	for _, v := range vs {
		vsid := v.ScopeID()
		if _, ok := s.data[vsid]; !ok {
			s.data[vsid] = make(map[sdktypes.Symbol]sdktypes.Var)
		}
		s.data[vsid][v.Name()] = v
	}
	return nil
}

func (s *fakeVarsService) Get(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error) {
	vs := sdktypes.NewVars()
	if len(names) == 0 {
		names = slices.Collect(maps.Keys(s.data[sid]))
	}
	for _, name := range names {
		if v, ok := s.data[sid][name]; ok {
			vs = vs.Append(v)
		}
	}
	return vs, nil
}

func (s *fakeVarsService) Delete(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) error {
	for _, name := range names {
		delete(s.data[sid], name)
	}
	return nil
}

func (s *fakeVarsService) FindActiveConnectionIDs(ctx context.Context, iid sdktypes.IntegrationID, name sdktypes.Symbol, value string) ([]sdktypes.ConnectionID, error) {
	return nil, errors.New("not implemented")
}
