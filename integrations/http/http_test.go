package http

import (
	"context"
	"encoding/json"
	"os"
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

var jsonContentHeader = map[string]string{contentTypeHeader: contentTypeJSON}

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

func TestParseBodyForRequest(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}       // body to warp as sdktypes.Value and parse
		headers        map[string]string // optional headers
		bodyType       string            // parsed body type
		reqBody        string            // body extracted from resulting request
		reqContentType string
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
			name:     "string + contentTypeJson => json",
			body:     "meow",
			headers:  jsonContentHeader,
			bodyType: bodyTypeJSON,
			reqBody:  `"meow"`,
		},
		{
			name:     "json string => raw",
			body:     `{"k":"v"}`,
			bodyType: bodyTypeRaw,
			reqBody:  `{"k":"v"}`,
		},
		{
			name:     "json string + contentTypeJson => json",
			body:     `{"k":"v"}`,
			headers:  jsonContentHeader,
			bodyType: bodyTypeJSON,
			reqBody:  `"{\"k\":\"v\"}"`,
		},
		{
			// NOTE: different behavior then python's requests.
			// Python lib will form encode map[string]interface{} as well
			name:           "dict (map[string]string) => form",
			body:           map[string]string{"k": "v"},
			bodyType:       bodyTypeForm,
			reqBody:        "k=v",
			reqContentType: contentTypeForm,
		},
		{
			// NOTE: different behavior then python's requests. See compatibility test
			// Pyhton lib will form encode although content-type is set to json
			name:           "dict (map[string]string) + contentTypeJSON => json",
			body:           map[string]string{"k": "v"},
			headers:        jsonContentHeader,
			bodyType:       bodyTypeJSON,
			reqBody:        `{"k":"v"}`,
			reqContentType: contentTypeJSON,
		},
		{
			name:           "dict (map[string]interface{}) => json",
			body:           map[string]interface{}{"k": "v", "t": true},
			headers:        jsonContentHeader,
			bodyType:       bodyTypeJSON,
			reqBody:        `{"k":"v","t":true}`,
			reqContentType: contentTypeJSON,
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
			if tt.reqContentType != "" {
				assert.Equal(t, tt.reqContentType, req.headers[contentTypeHeader])
			}
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
			args:   []interface{}{"meow"},
			errStr: "pass non-URL arguments as kwargs only",
		},
		{
			name:   "don't ignore data= in POST",
			method: "POST",
			// args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"data": "meow"},
			body:   "meow",
		},
		{
			name:   "ignore data= in GET",
			method: "GET",
			// args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"data": "meow"},
			body:   "",
		},
		{
			name:   "passing json",
			method: "POST",
			// args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"json": "woof"},
			body:   `"woof"`,
		},
		{
			name:   "passing json + body #1. json ignored",
			method: "POST",
			// args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"data": "meow", "json": "woof"},
			body:   "meow",
		},
		{
			name:   "passing json + body #2. json ignored",
			method: "POST",
			// args:   []interface{}{"http://dummy.url"},
			kwargs: map[string]interface{}{"json": "woof", "data": "meow"},
			body:   "meow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req request

			args := []interface{}{"http://dummy.url"}
			args = append(args, tt.args...)

			sdkArgs, err := kittehs.TransformError(args, sdktypes.WrapValue)
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

func TestPythonRequestsCompatibility(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping actual http requests test in CI")
	}

	type expected struct {
		data        string
		json        interface{}
		form        map[string]interface{}
		contentType string
	}

	sdkArgs := []sdktypes.Value{sdktypes.NewStringValue("http://httpbin.org/post")}
	method := "POST"
	nilForm := map[string]interface{}{}
	j1 := map[string]interface{}{"k": "v"}

	tests := []struct {
		name   string
		kwargs map[string]interface{}
		exp    expected
	}{
		{
			name:   "data = string",
			kwargs: map[string]interface{}{"data": "meow"},
			exp:    expected{data: "meow", form: nilForm, json: nil},
		},
		{
			name:   "data = not a json string",
			kwargs: map[string]interface{}{"data": `{'k':'v'}`},
			exp:    expected{data: `{'k':'v'}`, form: nilForm, json: nil},
		},
		{
			name:   "data = a json string. Apperar in data and in json",
			kwargs: map[string]interface{}{"data": `{"k":"v"}`},
			exp:    expected{data: "{\"k\":\"v\"}", form: nilForm, json: j1},
		},
		{
			name:   "data = dict. Form-encode the body and set the Content-Type",
			kwargs: map[string]interface{}{"data": j1},
			exp:    expected{data: "", form: j1, json: nil, contentType: contentTypeForm},
		},

		// NOTE: corner-case/library-dependent behavior?
		// Python's requests will first form encode dictionary passed via data=, even if
		// content-type = applicattion/json is specified. JS axious on the other hand will convert to json.
		// Meanwhile if content-type is explicitly set to JSON, we will convert to JSON as well
		// {
		// 	name:   "data = dict + content-type=json. Server will fail to recognize the form",
		// 	kwargs: map[string]interface{}{"data": jsonData, "headers": jsonContentHeader},
		// 	exp:    expected{data: "", form: jsonData, json: nil, contentType: contentTypeForm},
		// },

		{
			name:   "json = string => data=string + content-type",
			kwargs: map[string]interface{}{"json": "meow"},
			exp:    expected{data: `"meow"`, form: nilForm, json: `meow`, contentType: contentTypeJSON},
		},
		{
			name:   "json = `not a json string` => data=string + content-type",
			kwargs: map[string]interface{}{"json": `{'k':'v'}`},
			exp:    expected{data: `"{'k':'v'}"`, form: nilForm, json: `{'k':'v'}`, contentType: contentTypeJSON},
		},
		{
			// Although it's a vlid json string, it shouldn't be parsed as json.
			// json= should accept only valid json dict, not string. So passing json string should be treated as string
			name:   "json = `a json string` => data=string(!json) + content-type",
			kwargs: map[string]interface{}{"json": `{"k":"v"}`},
			exp:    expected{data: `"{\"k\":\"v\"}"`, form: nilForm, json: `{"k":"v"}`, contentType: contentTypeJSON},
		},

		// NOTE: python requests will escape the quotes in data. But this is not required
		{
			name:   "json = dict => data=json + json + content-type",
			kwargs: map[string]interface{}{"json": j1},
			exp:    expected{data: `{"k":"v"}`, form: nilForm, json: j1, contentType: contentTypeJSON},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req request
			sdkKwargs, err := kittehs.TransformMapValuesError(tt.kwargs, sdktypes.WrapValue)
			assert.NoError(t, err)

			err = unpackAndParseArgs(&req, method, sdkArgs, sdkKwargs)
			assert.NoError(t, err)

			resp, err := sendHttpRequest(context.Background(), req, method)
			assert.NoError(t, err)

			var respJson map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&respJson)
			assert.NoError(t, err)

			assert.Equal(t, tt.exp.data, respJson["data"])
			assert.Equal(t, tt.exp.json, respJson["json"])
			assert.Equal(t, tt.exp.form, respJson["form"])

			conentType, _ := respJson["headers"].(map[string]interface{})["Content-Type"].(string)
			assert.Equal(t, tt.exp.contentType, conentType)
		})
	}
}
