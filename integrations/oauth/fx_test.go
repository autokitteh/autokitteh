package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
		wantErr bool
	}{
		{
			name:    "empty",
			address: "",
			want:    defaultPublicBackendBaseURL,
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
			got, err := normalizeAddress(tt.address, defaultPublicBackendBaseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, got, tt.want)
		})
	}
}
