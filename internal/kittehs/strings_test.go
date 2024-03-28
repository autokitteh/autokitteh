package kittehs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	assert.Equal(t, ToString(String("meow")), "meow")
}

func TestPadLeft(t *testing.T) {
	assert.Equal(t, "01", PadLeft("1", '0', 2))
	assert.Equal(t, "1", PadLeft("1", '0', 1))
	assert.Equal(t, "000", PadLeft("", '0', 3))
	assert.Equal(t, "123", PadLeft("123", '0', 2))
}

func TestMatchLongetSuffix(t *testing.T) {
	assert.Equal(t, "", MatchLongestSuffix("", []string{"1", "3"}))
	assert.Equal(t, "234", MatchLongestSuffix("1234", []string{"4", "234", "34", "23"}))
}

func TestNormalizeURL(t *testing.T) {
	type args struct {
		rawURL string
		secure bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "basic happy path",
			args:    args{rawURL: "http://example.com", secure: false},
			want:    "http://example.com",
			wantErr: false,
		},
		{
			name:    "add HTTP scheme prefix",
			args:    args{rawURL: "example.com", secure: false},
			want:    "http://example.com",
			wantErr: false,
		},
		{
			name:    "add HTTPS scheme prefix",
			args:    args{rawURL: "example.com", secure: true},
			want:    "https://example.com",
			wantErr: false,
		},
		{
			name:    "change scheme",
			args:    args{rawURL: "http://example.com", secure: true},
			want:    "https://example.com",
			wantErr: false,
		},
		{
			name:    "strip path",
			args:    args{rawURL: "http://example.com/foo/bar", secure: false},
			want:    "http://example.com",
			wantErr: false,
		},
		{
			name:    "add scheme and strip path",
			args:    args{rawURL: "example.com/path", secure: true},
			want:    "https://example.com",
			wantErr: false,
		},
		{
			name:    "no host",
			args:    args{rawURL: "/path", secure: true},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeURL(tt.args.rawURL, tt.args.secure)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
