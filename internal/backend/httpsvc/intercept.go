package httpsvc

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
)

type ctxKey string

var t0CtxKey = ctxKey("t0")

func GetT0(ctx context.Context) time.Time {
	if t0, ok := ctx.Value(t0CtxKey).(time.Time); ok {
		return t0
	}

	return time.Time{}
}

type responseInterceptor struct {
	http.ResponseWriter
	StatusCode int
	logger     *zap.Logger
}

func (r *responseInterceptor) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseInterceptor) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
		return
	}

	r.logger.Error(
		"responseInterceptor: underlying http.ResponseWriter does not implement http.Flusher",
		zap.String("type", fmt.Sprintf("%T", r.ResponseWriter)),
	)
}

type RequestLogExtractor func(*http.Request) []zap.Field

func intercept(z *zap.Logger, cfg *LoggerConfig, extractors []RequestLogExtractor, next http.Handler) (http.HandlerFunc, error) {
	// MustCompile is appropriate here because the patterns are static
	// and errors in them should be caught at startup. Furthermore, we
	// combine them into a single regular expression, because efficiency is
	// critical when we delay each and every incoming HTTP and gRPC request.
	unimportant := regexp.MustCompile(strings.Join(cfg.UnimportantRegexes, `|`))
	unlogged := regexp.MustCompile(strings.Join(cfg.UnloggedRegexes, `|`))

	return func(w http.ResponseWriter, r *http.Request) {
		if unlogged.MatchString(r.URL.Path) {
			return
		}

		w.Header().Set("X-AutoKitteh-Process-ID", fixtures.ProcessID())

		z := z.With(zap.String("method", r.Method), zap.String("path", r.URL.Path))
		msg := fmt.Sprintf("HTTP request: %s %s", r.Method, r.URL.Path)

		level := cfg.ImportantLevel.Level()
		if unimportant.MatchString(r.URL.Path) {
			level = cfg.UnimportantLevel.Level()
		}

		if ce := z.Check(level, msg); ce != nil {
			var fields []zap.Field
			for _, x := range extractors {
				fields = append(fields, x(r)...)
			}

			ce.Write(fields...)
		}

		rwi := &responseInterceptor{ResponseWriter: w, StatusCode: http.StatusOK, logger: z}

		w.Header().Set("Trailer", "X-AutoKitteh-Duration")

		t0 := time.Now()
		r = r.WithContext(context.WithValue(r.Context(), t0CtxKey, t0))

		next.ServeHTTP(rwi, r)
		d := time.Since(t0)

		w.Header().Add("X-AutoKitteh-Duration", d.Truncate(time.Microsecond).String())

		if rwi.StatusCode >= 400 {
			level = cfg.ErrorsLevel.Level()
		}

		msg = strings.Replace(msg, "request", "response", 1)
		if ce := z.Check(level, msg); ce != nil {
			ce.Write(zap.Int("statusCode", rwi.StatusCode), zap.Duration("duration", d))
		}
	}, nil
}
