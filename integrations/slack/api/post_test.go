package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type response struct {
	OK               bool              `json:"ok"`
	Response         string            `json:"response,omitempty"`
	Warning          string            `json:"warning,omitempty"`
	Error            string            `json:"error,omitempty"`
	ResponseMetadata *ResponseMetadata `json:"response_metadata,omitempty"`
}

func TestPostForm(t *testing.T) {
	tests := []struct {
		name        string
		kv          url.Values
		startServer bool
		respBody    []byte
		wantErr     bool
		wantResp    string
	}{
		{
			name:        "happy_path",
			startServer: true,
			respBody:    []byte(`{"ok": true, "response": "response"}`),
			wantResp:    "response",
		},
		{
			name:        "bad_response",
			startServer: true,
			respBody:    []byte("bad"),
			wantErr:     true,
		},
		{
			name:        "not_ok_without_metadata",
			startServer: true,
			respBody:    []byte(`{"ok": false, "error": "error"}`),
		},
		{
			name:        "not_ok_with_metadata",
			startServer: true,
			respBody:    []byte(`{"ok": false, "error": "error", "response_metadata": {"messages": ["msg"]}}`),
		},
		{
			name:        "warning",
			startServer: true,
			respBody:    []byte(`{"ok": true, "warning": "warning"}`),
		},
		{
			name:        "server_not_responding",
			startServer: false,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, string(tt.respBody))
				got := r.Header.Get("Authorization")
				want := ""
				if got != want {
					t.Errorf("PostForm() Authorization header = %q, want %q", got, want)
				}
			})

			s := httptest.NewUnstartedServer(handler)
			if tt.startServer {
				s.Start()
			}
			defer s.Close()

			slackURL = s.URL + "/"
			got := &response{}
			err := PostForm(context.Background(), nil, "test", tt.kv, got, "slack.method")
			if (err != nil) != tt.wantErr {
				t.Errorf("PostForm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantResp != "" && got.Response != tt.wantResp {
				t.Errorf("PostForm() response = %v, want %q", got, tt.wantResp)
			}
		})
	}
}

func TestPostJSON(t *testing.T) {
	tests := []struct {
		name        string
		json        []byte
		startServer bool
		respBody    []byte
		wantErr     bool
		wantResp    string
	}{
		{
			name:        "happy_path",
			startServer: true,
			respBody:    []byte(`{"ok": true, "response": "response"}`),
			wantResp:    "response",
		},
		{
			name:        "bad_response",
			startServer: true,
			respBody:    []byte("bad"),
			wantErr:     true,
		},
		{
			name:        "not_ok",
			startServer: true,
			respBody:    []byte(`{"ok": false}`),
		},
		{
			name:        "warning",
			startServer: true,
			respBody:    []byte(`{"ok": true, "warning": "warning"}`),
		},
		{
			name:        "server_not_responding",
			startServer: false,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, string(tt.respBody))
				got := r.Header.Get("Authorization")
				want := ""
				if got != want {
					t.Errorf("PostForm() Authorization header = %q, want %q", got, want)
				}
			})

			s := httptest.NewUnstartedServer(handler)
			if tt.startServer {
				s.Start()
			}
			defer s.Close()

			slackURL = s.URL + "/"
			got := &response{}
			err := PostJSON(context.Background(), nil, "test", tt.json, got, "slack.method")
			if (err != nil) != tt.wantErr {
				t.Errorf("PostJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantResp != "" && got.Response != tt.wantResp {
				t.Errorf("PostForm() response = %v, want %q", got, tt.wantResp)
			}
		})
	}
}
