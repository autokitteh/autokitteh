package http

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func TestExtractPathKeys(t *testing.T) {
	tests := []struct {
		path string
		want []string
	}{
		{},
		{
			path: "/a",
		},
		{
			path: "/a/{b}/c",
			want: []string{"b"},
		},
		{
			path: "/a/{b...}",
			want: []string{"b"},
		},
		{
			path: "/{meow}/moo/{woof}/x",
			want: []string{"meow", "woof"},
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			got, err := extractPathKeys(test.path)
			if assert.NoError(t, err) {
				return
			}

			assert.Equal(t, test.want, got)
		})
	}

	_, err := extractPathKeys("{")
	assert.Error(t, err)

	_, err = extractPathKeys("a/{}")
	assert.Error(t, err)

	_, err = extractPathKeys("a/{...}")
	assert.Error(t, err)
}

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
