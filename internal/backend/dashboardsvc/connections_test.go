package dashboardsvc

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_initResult(t *testing.T) {
	basePath := "/connections/cid_12345/result"

	tests := []struct {
		name   string
		path   string
		status int
		err    string
	}{
		{
			name:   "no status",
			path:   basePath,
			status: 500,
			err:    "non-integer status",
		},
		{
			name:   "invalid status",
			path:   basePath + "?status=abc",
			status: 500,
			err:    "non-integer status",
		},
		{
			name:   "happy path for init success",
			path:   basePath + "?status=200",
			status: 200,
			err:    "",
		},
		{
			name:   "happy path for init failure",
			path:   basePath + "?status=500&error=oh%20no",
			status: 500,
			err:    "oh no",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", basePath+tt.path, nil)
			w := httptest.NewRecorder()
			initResult(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			assert.Contains(t, string(body), fmt.Sprintf(`"status":%d`, tt.status))
			if tt.err == "" {
				assert.NotContains(t, string(body), `"error":`)
			} else {
				assert.Contains(t, string(body), fmt.Sprintf(`"error":"%s"`, tt.err))
			}
		})
	}
}
