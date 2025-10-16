package common

import (
	"context"
	"errors"
	"maps"
	"slices"
	"testing"
	"time"

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

func TestOAuthDataToToken(t *testing.T) {
	tests := []struct {
		name   string
		expiry string
		want   time.Time
	}{
		{
			name:   "no_expiry",
			expiry: "",
			want:   time.Time{},
		},
		{
			name:   "explicit_zero_expiry",
			expiry: "0001-01-01T00:00:00Z",
			want:   time.Time{},
		},
		{
			name:   "valid_expiry",
			expiry: "2012-12-12T00:00:00Z",
			want:   time.Date(2012, 12, 12, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := OAuthData{Expiry: tt.expiry}
			if got := o.ToToken().Expiry; got != tt.want {
				t.Errorf("OAuthData.ToToken().Expiry = %v, want %v", got, tt.want)
			}
		})
	}
}
