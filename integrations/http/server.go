package http

import (
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const routePrefix = "/http/"

// HTTPHandler is an autokitteh webhook which implements [http.Handler] to
// receive and dispatch asynchronous event notifications.
type HTTPHandler struct {
	dispatcher sdkservices.Dispatcher
	logger     *zap.Logger
}

func Start(l *zap.Logger, mux *http.ServeMux, d sdkservices.Dispatcher) {
	h := HTTPHandler{dispatcher: d, logger: l}
	mux.Handle(routePrefix, http.StripPrefix(routePrefix, h))
}

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("url", r.URL.String()))

	l.Info("Incoming request")

	// TODO: Use a real connection token.
	token, _, _ := strings.Cut(r.URL.Path, "/")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error("body read error", zap.Error(err))
		// no return
	}

	data := map[string]sdktypes.Value{
		"url": kittehs.Must1(sdktypes.NewStructValue(
			sdktypes.NewStringValue("url"),
			map[string]sdktypes.Value{
				"scheme":       sdktypes.NewStringValue(r.URL.Scheme),
				"opaque":       sdktypes.NewStringValue(r.URL.Opaque),
				"host":         sdktypes.NewStringValue(r.URL.Host),
				"fragment":     sdktypes.NewStringValue(r.URL.Fragment),
				"raw_fragment": sdktypes.NewStringValue(r.URL.RawFragment),
				"raw":          sdktypes.NewStringValue(r.URL.RawPath),
				"path":         sdktypes.NewStringValue(r.URL.Path),
				"raw_query":    sdktypes.NewStringValue(r.URL.RawQuery),
				"query": sdktypes.NewDictValueFromStringMap(
					kittehs.TransformMapValues(r.URL.Query(), func(vs []string) sdktypes.Value {
						return sdktypes.NewStringValue(strings.Join(vs, ","))
					}),
				),
			},
		)),
		"method": sdktypes.NewStringValue(r.Method),
		"header": sdktypes.NewDictValueFromStringMap(
			kittehs.TransformMapValues(r.Header, func(vs []string) sdktypes.Value {
				return sdktypes.NewStringValue(strings.Join(vs, ","))
			}),
		),
		// TODO(ENG-294): return an object that can has various decoding methods.
		"body_bytes":  sdktypes.NewBytesValue(body),
		"body_string": sdktypes.NewStringValue(string(body)),
		"body":        sdktypes.NewStringValue(string(body)),
	}

	event, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		IntegrationId:    integrationID.String(),
		IntegrationToken: token,
		EventType:        strings.ToLower(r.Method),
		Data:             kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("create event error", zap.Error(err))
		return
	}

	eid, err := h.dispatcher.Dispatch(r.Context(), event, nil)
	if err != nil {
		l.Error("dispatch error", zap.Error(err))
	}

	l.Info("dispatched", zap.String("event_id", eid.String()))
}
