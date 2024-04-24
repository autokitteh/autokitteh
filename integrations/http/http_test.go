package http

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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

func TestParseBody(t *testing.T) {
	tests := []struct {
		name     string
		body     interface{}       // body to warp as sdktypes.Value and parse
		headers  map[string]string // optional headers
		bodyType string            // parsed body type
		reqBody  string            // body extracted from resulting request
	}{
		{
			name:     "empty body",
			body:     "",
			bodyType: bodyTypeRaw,
			reqBody:  "",
		},
		{
			name:     "string",
			body:     "meow",
			bodyType: bodyTypeRaw,
			reqBody:  "meow",
		},
		{
			name:     "json as string => raw",
			body:     `{"k":"v"}`,
			bodyType: bodyTypeRaw,
			reqBody:  `{"k":"v"}`,
		},
		{
			name:     "map[string]string => form",
			body:     map[string]string{"k": "v"},
			bodyType: bodyTypeForm,
			reqBody:  "k=v",
		},
		{
			name:     "map[string]string + contentTypeJSON => json",
			body:     map[string]string{"k": "v"},
			headers:  map[string]string{contentTypeHeader: contentTypeJSON},
			bodyType: bodyTypeJSON,
			reqBody:  `{"k":"v"}`,
		},
		{
			name:     "unmarshal(json string) => json",
			body:     map[string]interface{}{"k": "v", "t": true},
			headers:  map[string]string{contentTypeHeader: contentTypeJSON},
			bodyType: bodyTypeJSON,
			reqBody:  `{"k":"v","t":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawBody string
			var jsonBody sdktypes.Value
			var formBody map[string]string

			bodyToParse, err := sdktypes.WrapValue(tt.body) // warp into sdktypes.Value
			assert.NoError(t, err)
			fmt.Printf("%v\n", bodyToParse)

			err, bodyType := parseBody(bodyToParse, tt.headers, &rawBody, &formBody, &jsonBody)
			assert.NoError(t, err)
			req, err := http.NewRequest("POST", "http://dummy.url", nil) // create dummy request
			assert.NoError(t, err)

			var body []byte = nil
			err = setBody(req, bodyType, rawBody, formBody, jsonBody)
			assert.NoError(t, err)
			assert.Equal(t, tt.bodyType, bodyType)

			if req.Body != nil {
				body, err = io.ReadAll(req.Body)
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.reqBody, string(body))
		})
	}
}
