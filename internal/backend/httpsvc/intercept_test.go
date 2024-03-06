package httpsvc

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// mockWriter implements http.ResponseWriter but not http.Flusher
type mockWriter struct{}

func (m mockWriter) Header() http.Header {
	return nil
}

func (m mockWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (m mockWriter) WriteHeader(statusCode int) {}

func Test_responseInterceptor_Flusher(t *testing.T) {
	var buf bytes.Buffer
	logger := zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zapcore.EncoderConfig{}),
			zapcore.AddSync(&buf),
			zapcore.InfoLevel,
		),
		zap.ErrorOutput(zapcore.AddSync(&buf)),
	)
	w := httptest.NewRecorder()

	var r http.ResponseWriter = &responseInterceptor{w, 0, logger}
	_, ok := r.(http.Flusher)
	require.True(t, ok)

	buf.Reset()
	r = &responseInterceptor{mockWriter{}, 0, logger}
	f, ok := r.(http.Flusher)
	require.True(t, ok)
	f.Flush()
	require.Greater(t, len(buf.String()), 0)
}
