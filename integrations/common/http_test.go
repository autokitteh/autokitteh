package common

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHTTPGet(t *testing.T) {
	payload := `{"foo": "bar"}`

	// Create a fake HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		assert.Equal(t, r.Header.Get(HeaderAccept), ContentTypeJSON)
		assert.Equal(t, r.Header.Get(HeaderContentType), "")

		body, err := io.ReadAll(r.Body)
		assert.NilError(t, err)
		assert.Equal(t, len(body), 0)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(payload))
		assert.NilError(t, err)
	}))
	t.Cleanup(server.Close)

	// Call the function under test.
	ctx := t.Context()
	resp, err := HTTPGet(ctx, server.URL, "")
	assert.NilError(t, err)

	// Check the response.
	assert.Equal(t, string(resp), payload)
}

func TestHTTPPostForm(t *testing.T) {
	payload := `{"foo": "bar"}`

	// Create a fake HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		assert.Equal(t, r.Header.Get(HeaderAccept), ContentTypeJSON)
		assert.Equal(t, r.Header.Get(HeaderContentType), ContentTypeForm)

		assert.NilError(t, r.ParseForm())
		assert.Equal(t, r.FormValue("key"), "value")

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(payload))
		assert.NilError(t, err)
	}))
	t.Cleanup(server.Close)

	// Prepare the test request's payload.
	req := url.Values{}
	req.Set("key", "value")

	// Call the function under test.
	ctx := t.Context()
	resp, err := HTTPPostForm(ctx, server.URL, "", req)
	assert.NilError(t, err)

	// Check the response.
	assert.Equal(t, string(resp), payload)
}

func TestHTTPPostJSON(t *testing.T) {
	payload := "{}"

	// Create a fake HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		assert.Equal(t, r.Header.Get(HeaderAccept), ContentTypeJSON)

		body, err := io.ReadAll(r.Body)
		assert.NilError(t, err)
		if len(body) > 0 {
			assert.Equal(t, r.Header.Get(HeaderContentType), ContentTypeJSONCharsetUTF8)
			assert.Equal(t, string(body), payload)
		} else {
			assert.Equal(t, r.Header.Get(HeaderContentType), "")
		}

		if r.Header.Get(HeaderAuthorization) != "" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
		_, err = w.Write([]byte(payload))
		assert.NilError(t, err)
	}))
	t.Cleanup(server.Close)

	// Test cases.
	tests := []struct {
		name    string
		auth    string
		req     any
		wantErr bool
	}{
		{
			name: "with_auth_string_body_ok",
			auth: "auth",
			req:  payload,
		},
		{
			name: "with_auth_map_body_ok",
			auth: "auth",
			req:  map[string]any{},
		},
		{
			name:    "without_auth_without_body_err",
			auth:    "",
			req:     http.NoBody,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function under test.
			ctx := t.Context()
			resp, err := HTTPPostJSON(ctx, server.URL, tt.auth, tt.req)

			// Check the response.
			if !tt.wantErr {
				assert.NilError(t, err)
				assert.Equal(t, string(resp), payload)
			} else {
				assert.Error(t, err, "401 Unauthorized: {}")
			}
		})
	}
}
