package microsoft

import (
	"net/url"
	"reflect"
	"testing"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestMapToVars(t *testing.T) {
	cid := sdktypes.NewConnectionID()
	tests := []struct {
		name string
		m    map[sdktypes.Symbol]string
		vsid sdktypes.VarScopeID
		want sdktypes.Vars
	}{
		{
			name: "empty",
			m:    map[sdktypes.Symbol]string{},
			vsid: sdktypes.InvalidVarScopeID,
			want: sdktypes.NewVars(),
		},
		{
			name: "nonempty",
			m: map[sdktypes.Symbol]string{
				clientIDVar:     "XXX",
				clientSecretVar: "YYY",
			},
			vsid: sdktypes.NewVarScopeID(cid),
			want: sdktypes.NewVars().Append(
				sdktypes.NewVar(clientIDVar).WithScopeID(sdktypes.NewVarScopeID(cid)).SetValue("XXX").SetSecret(false),
				sdktypes.NewVar(clientSecretVar).WithScopeID(sdktypes.NewVarScopeID(cid)).SetValue("YYY").SetSecret(true),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapToVars(tt.m, tt.vsid); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapToVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOAuthURL(t *testing.T) {
	tests := []struct {
		name   string
		cid    string
		origin string
		vs     url.Values
		want   string
	}{
		{
			name:   "all_scopes",
			cid:    "XXX",
			origin: "YYY",
			vs:     url.Values{},
			want:   "/oauth/start/microsoft?cid=XXX&origin=YYY",
		},
		{
			name:   "narrow_scopes",
			cid:    "XXX",
			origin: "YYY",
			vs:     map[string][]string{"auth_scopes": {"ZZZ"}},
			want:   "/oauth/start/microsoft-ZZZ?cid=XXX&origin=YYY",
		},
		{
			name:   "malicious_scopes",
			cid:    "XXX",
			origin: "YYY",
			vs:     map[string][]string{"auth_scopes": {"ZZZ/.."}},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := sdkintegrations.ConnectionInit{
				ConnectionID: tt.cid,
				Origin:       tt.origin,
			}
			if got := oauthURL(tt.vs, ci); got != tt.want {
				t.Errorf("oauthURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
