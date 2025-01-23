package microsoft

import (
	"net/url"
	"testing"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
)

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
