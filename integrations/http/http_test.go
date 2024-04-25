package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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
			var req request
			var body string

			bodyToParse, err := sdktypes.WrapValue(tt.body) // warp into sdktypes.Value
			assert.NoError(t, err)

			if tt.headers != nil {
				req.headers = tt.headers
			} else {
				req.headers = make(map[string]string)
			}
			err = parseBody(&req, bodyToParse)
			assert.NoError(t, err)
			assert.Equal(t, tt.bodyType, req.bodyType)

			if req.body != nil {
				body = req.body.String()
			}
			assert.Equal(t, tt.reqBody, body)
		})
	}
}

func TestUnpackAndParseArgs(t *testing.T) {
	tests := []struct {
		name   string
		method string
		args   []interface{}
		kwargs map[string]interface{}
		errStr string
		body   string
	}{
		{
			name:   "disallow any args except URL",
			method: "GET",
			args:   []interface{}{"http://dummy.url", "meow"},
			errStr: "pass non-URL arguments as kwargs only",
		},
		{
			name:   "don't ignore data= in POST",
			method: "POST",
			args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"data": "meow"},
			body:   "meow",
		},
		{
			name:   "ignore data= in GET",
			method: "GET",
			args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"data": "meow"},
			body:   "",
		},
		{
			name:   "passing json",
			method: "POST",
			args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"json": "woof"},
			body:   "woof",
		},
		{
			name:   "passing json + body #1. json ignored",
			method: "POST",
			args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"data": "meow", "json": "woof"},
			body:   "meow",
		},
		{
			name:   "passing json + body #2. json ignored",
			method: "POST",
			args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"json": "woof", "data": "meow"},
			body:   "meow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req request

			sdkArgs, err := kittehs.TransformError(tt.args, sdktypes.WrapValue)
			assert.NoError(t, err)
			sdkKwargs, err := kittehs.TransformMapValuesError(tt.kwargs, sdktypes.WrapValue)
			assert.NoError(t, err)

			err = unpackAndParseArgs(&req, tt.method, sdkArgs, sdkKwargs)
			if tt.errStr != "" {
				assert.ErrorContains(t, err, tt.errStr)
				return
			} else {
				assert.NoError(t, err)
			}

			reqBody := ""
			if req.body != nil {
				reqBody = req.body.String()
			}
			assert.Equal(t, tt.body, reqBody)
		})
	}
}
