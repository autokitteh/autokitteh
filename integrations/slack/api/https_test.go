package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type httpTestResponse struct {
	SlackResponse

	Response string `json:"response,omitempty"`
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		startServer bool
		respBody    string
		wantErr     bool
		wantResp    string
	}{
		{
			name:        "happy_path",
			startServer: true,
			respBody:    `{"ok": true, "response": "response"}`,
			wantResp:    "response",
		},
		{
			name:        "bad_response",
			startServer: true,
			respBody:    "bad",
			wantErr:     true,
		},
		{
			name:        "slack_not_ok",
			startServer: true,
			respBody:    `{"ok": false}`,
		},
		{
			name:    "server_not_responding",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewUnstartedServer(handler(t, tt.respBody))
			if tt.startServer {
				s.Start()
			}
			defer s.Close()

			slackURL = s.URL
			got := &httpTestResponse{}
			err := get(context.TODO(), "", "slack.method", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantResp != "" && got.Response != tt.wantResp {
				t.Errorf("get() response = %v, want %q", got, tt.wantResp)
			}
		})
	}
}

func TestPost(t *testing.T) {
	tests := []struct {
		name        string
		startServer bool
		respBody    string
		wantErr     bool
		wantResp    string
	}{
		{
			name:        "happy_path",
			startServer: true,
			respBody:    `{"ok": true, "response": "response"}`,
			wantResp:    "response",
		},
		{
			name:        "bad_response",
			startServer: true,
			respBody:    "bad",
			wantErr:     true,
		},
		{
			name:        "slack_not_ok",
			startServer: true,
			respBody:    `{"ok": false}`,
		},
		{
			name:    "server_not_responding",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewUnstartedServer(handler(t, tt.respBody))
			if tt.startServer {
				s.Start()
			}
			defer s.Close()

			slackURL = s.URL
			got := &httpTestResponse{}
			err := Post(context.TODO(), "", "slack.method", nil, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantResp != "" && got.Response != tt.wantResp {
				t.Errorf("Post() response = %v, want %q", got, tt.wantResp)
			}
		})
	}
}

func handler(t *testing.T, resp string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		want := ""
		if got != want {
			t.Errorf("authorization header = %q, want %q", got, want)
		}

		fmt.Fprint(w, resp)
	})
}
