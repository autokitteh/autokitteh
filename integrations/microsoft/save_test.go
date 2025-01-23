package microsoft

import (
	"testing"
)

func TestOAuthURLX(t *testing.T) {
	tests := []struct {
		name   string
		cid    string
		origin string
		scopes string
		want   string
		err    bool
	}{
		{
			name:   "all_scopes",
			cid:    "XXX",
			origin: "YYY",
			scopes: "",
			want:   "/oauth/start/microsoft?cid=XXX&origin=YYY",
		},
		{
			name:   "narrow_scopes",
			cid:    "XXX",
			origin: "YYY",
			scopes: "ZZZ",
			want:   "/oauth/start/microsoft-ZZZ?cid=XXX&origin=YYY",
		},
		{
			name:   "malicious_scopes",
			cid:    "XXX",
			origin: "YYY",
			scopes: "ZZZ/..",
			err:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := oauthURL(tt.cid, tt.origin, tt.scopes)
			if (err != nil) != tt.err {
				t.Errorf("oauthURL() error = %v, want error = %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("oauthURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
