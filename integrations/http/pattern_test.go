package http

import (
	"testing"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func TestMatchPattern(t *testing.T) {
	// Test cases
	tests := []struct {
		pattern string
		req     string
		want    map[string]string
		wantErr error
	}{
		{
			pattern: "/a",
			req:     "/b",
			wantErr: sdkerrors.ErrNotFound,
		},
		{
			pattern: "/a",
			req:     "/a",
		},
		{
			pattern: "/a/{b}",
			req:     "/a/c",
			want:    map[string]string{"b": "c"},
		},
		{
			pattern: "/a/{b}/{r}",
			req:     "/a/b/bla",
			want:    map[string]string{"b": "c", "r": "bla"},
		},
	}

	for _, test := range tests {
		t.Run(test.pattern+"|"+test.req, func(t *testing.T) {
			got, gotErr := MatchPattern(test.pattern, test.req)
			if gotErr != test.wantErr {
				t.Errorf("MatchPattern() error = %v, wantErr %v", gotErr, test.wantErr)
			}
			if len(got) != len(test.want) {
				t.Errorf("MatchPattern() got = %v, want %v", got, test.want)
			}
		})
	}
}
