package oauth

import (
	"testing"
)

func TestOAuthNormalizeAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
		wantErr bool
	}{
		{
			name:    "empty",
			address: "",
			want:    "https://address-not-configured",
		},
		{
			name:    "address_only",
			address: "example.com",
			want:    "https://example.com",
		},
		{
			name:    "http_prefix",
			address: "http://example.com",
			want:    "https://example.com",
		},
		{
			name:    "https_prefix",
			address: "https://example.com",
			want:    "https://example.com",
		},
		{
			name:    "ignore_suffixes",
			address: "example.com/path/to/somewhere?query=string&key=value#fragment",
			want:    "https://example.com",
		},
		{
			name:    "invalid",
			address: "://*&^%$",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OAuth{cfg: &Config{Address: tt.address}}
			if err := o.normalizeAddress(); (err != nil) != tt.wantErr {
				t.Errorf("normalizeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if o.BaseURL != tt.want {
				t.Errorf("cfg.Address = %q, want %q", o.BaseURL, tt.want)
			}
		})
	}
}
