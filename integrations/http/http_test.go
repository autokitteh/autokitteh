package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBodyToStructJSON(t *testing.T) {
	v := bodyToStruct([]byte(`{"i": 1, "f": 1.2}`), nil)
	require.True(t, v.IsStruct())

	json := v.GetStruct().Fields()["json"]
	require.True(t, json.IsFunction())

	data, err := json.GetFunction().ConstValue()
	require.NoError(t, err)

	require.True(t, data.IsDict())

	fs, err := data.GetDict().ToStringValuesMap()
	require.NoError(t, err)

	i := fs["i"]
	if assert.True(t, i.IsInteger()) {
		assert.Equal(t, int64(1), i.GetInteger().Value())
	}

	f := fs["f"]
	if assert.True(t, f.IsFloat()) {
		assert.Equal(t, 1.2, f.GetFloat().Value())
	}
}

func TestSetQueryParams(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		params  map[string]string
		wantURL string
		wantErr bool
	}{
		{
			name:    "empty params",
			rawURL:  "http://example.com",
			params:  map[string]string{},
			wantURL: "http://example.com",
			wantErr: false,
		},
		{
			name:    "single param",
			rawURL:  "http://example.com",
			params:  map[string]string{"a": "1"},
			wantURL: "http://example.com?a=1",
			wantErr: false,
		},
		{
			name:    "multiple params",
			rawURL:  "http://example.com",
			params:  map[string]string{"a": "1", "b": "2"},
			wantURL: "http://example.com?a=1&b=2",
			wantErr: false,
		},
		{
			name:    "add to existing params",
			rawURL:  "http://example.com?c=3&d=4",
			params:  map[string]string{"a": "1", "b": "2&"},
			wantURL: "http://example.com?a=1&b=2%26&c=3&d=4",
			wantErr: false,
		},
		{
			name:    "replace existing param",
			rawURL:  "http://example.com?c=2&d=4",
			params:  map[string]string{"a": "1", "c": "3"},
			wantURL: "http://example.com?a=1&c=3&d=4",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setQueryParams(&tt.rawURL, tt.params); (err != nil) != tt.wantErr {
				t.Errorf("setQueryParams() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.rawURL != tt.wantURL {
				t.Errorf("setQueryParams() got URL %q, want %q", tt.rawURL, tt.wantURL)
			}
		})
	}
}
