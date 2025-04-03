package oauth

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestAuthCodes(t *testing.T) {
	tests := []struct {
		name string
		opts map[string]string
		want []oauth2.AuthCodeOption
	}{
		{
			name: "empty",
			opts: nil,
			want: nil,
		},
		{
			name: "one",
			opts: map[string]string{
				"access_type": "offline",
			},
			want: []oauth2.AuthCodeOption{
				oauth2.AccessTypeOffline,
			},
		},
		{
			name: "two",
			opts: map[string]string{
				"access_type": "offline",
				"prompt":      "consent",
			},
			want: []oauth2.AuthCodeOption{
				oauth2.AccessTypeOffline,
				oauth2.ApprovalForce,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := authCodes(tt.opts)
			assert.Len(t, got, len(tt.want))
		})
	}
}

func TestGuessFrontendURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "webhooks_address_not_configured",
			baseURL: defaultPublicBackendBaseURL,
			want:    "https://app.autokitteh.cloud",
		},
		{
			name:    "multi_tenant_cloud",
			baseURL: "https://api.autokitteh.cloud",
			want:    "https://app.autokitteh.cloud",
		},
		{
			name:    "single_tenant_cloud", // a.k.a. named customers
			baseURL: "https://customer-api.autokitteh.cloud",
			want:    "https://customer.autokitteh.cloud",
		},
		{
			name:    "self_hosted",
			baseURL: "http://autokitteh.ngrok.dev",
			want:    "https://app.autokitteh.cloud",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := guessFrontendURL(tt.baseURL)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseStateParam(t *testing.T) {
	tests := []struct {
		name       string
		state      string
		wantCID    sdktypes.ConnectionID
		wantOrigin string
		wantErr    bool
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "no_cid",
			state:   "_origin",
			wantErr: true,
		},
		{
			name:    "no_origin",
			state:   "01jpgcggdqe5htw595bxhj8d3p_",
			wantErr: true,
		},
		{
			name:    "no_separator",
			state:   "01jpgv1mqqe98rz5byx9a6fsc7origin",
			wantErr: true,
		},
		{
			name:       "valid",
			state:      "01jp6e37hhf5b9xyjcawbpgv7b_origin",
			wantCID:    kittehs.Must1(sdktypes.ParseConnectionID("con_01jp6e37hhf5b9xyjcawbpgv7b")),
			wantOrigin: "origin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCID, gotOrigin, err := parseStateParam(tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStateParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCID, tt.wantCID) {
				t.Errorf("parseStateParam() got CID = %v, want %v", gotCID, tt.wantCID)
			}
			if gotOrigin != tt.wantOrigin {
				t.Errorf("parseStateParam() got origin = %v, want %v", gotOrigin, tt.wantOrigin)
			}
		})
	}
}

func TestExtraData(t *testing.T) {
	tests := []struct {
		name  string
		extra map[string]any
		want  map[string]any
	}{
		{
			name:  "empty",
			extra: nil,
			want:  map[string]any{},
		},
		{
			name: "instance_url",
			extra: map[string]any{
				"instance_url": "https://example.com",
			},
			want: map[string]any{
				"instance_url": "https://example.com",
			},
		},
		{
			name: "instance_url_and_other",
			extra: map[string]any{
				"instance_url": "https://example.com",
				"other":        "value",
			},
			want: map[string]any{
				"instance_url": "https://example.com",
			},
		},
		{
			name: "other",
			extra: map[string]any{
				"other": "value",
			},
			want: map[string]any{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &oauth2.Token{}
			if tt.extra != nil {
				token = token.WithExtra(tt.extra)
			}
			if got := extraData(token); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extraData() = %v, want %v", got, tt.want)
			}
		})
	}
}
