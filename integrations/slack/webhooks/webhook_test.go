package webhooks

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"testing/iotest"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
)

func TestWebhookCheckRequest(t *testing.T) {
	tests := []struct {
		name            string
		gotContentType  string
		wantContentType string
		timestampHeader string
		signatureHeader string
		r               io.Reader
		want            []byte
	}{
		{
			name:            "wrong_Content-Type",
			gotContentType:  common.ContentTypeForm,
			wantContentType: common.ContentTypeJSON,
			timestampHeader: strconv.FormatInt(time.Now().Unix(), 10),
			signatureHeader: "v0=test",
			r:               nil,
			want:            nil,
		},
		{
			name:            "missing_X-Slack-Request-Timestamp",
			gotContentType:  common.ContentTypeJSON,
			wantContentType: common.ContentTypeJSON,
			timestampHeader: "",
			signatureHeader: "v0=test",
			r:               nil,
			want:            nil,
		},
		{
			name:            "non-numeric_X-Slack-Request-Timestamp",
			gotContentType:  common.ContentTypeForm,
			wantContentType: common.ContentTypeForm,
			timestampHeader: "abc",
			signatureHeader: "v0=test",
			r:               nil,
			want:            nil,
		},
		{
			name:            "X-Slack-Request-Timestamp_in_the_past",
			gotContentType:  common.ContentTypeForm,
			wantContentType: common.ContentTypeForm,
			timestampHeader: strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10),
			signatureHeader: "v0=test",
			r:               nil,
			want:            nil,
		},
		{
			name:            "X-Slack-Request-Timestamp_in_the_future",
			gotContentType:  common.ContentTypeForm,
			wantContentType: common.ContentTypeForm,
			timestampHeader: strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
			signatureHeader: "v0=test",
			r:               nil,
			want:            nil,
		},
		{
			name:            "missing_X-Slack-Signature",
			gotContentType:  common.ContentTypeForm,
			wantContentType: common.ContentTypeForm,
			timestampHeader: strconv.FormatInt(time.Now().Unix(), 10),
			signatureHeader: "",
			r:               nil,
			want:            nil,
		},
		{
			name:            "body_reader_error",
			gotContentType:  common.ContentTypeForm,
			wantContentType: common.ContentTypeForm,
			timestampHeader: strconv.FormatInt(time.Now().Unix(), 10),
			signatureHeader: "v0=test",
			r:               iotest.ErrReader(errors.New("test error")),
			want:            nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/test", tt.r)
			r.Header.Add(common.HeaderContentType, tt.gotContentType)
			r.Header.Add(headerSlackTimestamp, tt.timestampHeader)
			r.Header.Add(headerSlackSignature, tt.signatureHeader)

			h := handler{}
			got := h.checkRequest(w, r, zap.L(), tt.wantContentType)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("unexpected checkRequest() return value (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWebhookVerifySignature(t *testing.T) {
	timestamp := "123456"
	body := []byte("body")
	wantSignature := "v0=913933f8f8e04ae6fe0f66ec8fe5e548fdd2461ffe1175d9440377832e1b7f3b"

	if !verifySignature("", timestamp, wantSignature, body) {
		t.Errorf("verifySignature() = false, want true")
	}
}
