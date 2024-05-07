package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// Save new autokitteh connections with user-submitted secrets.
	uiPath = "/i/http/connect/"

	// savePath is the URL path for our handler to save a new autokitteh
	// connection, after the user submits its details via a web form.
	savePath = "/i/http/save"
)

// handler is an autokitteh webhook which implements [http.Handler] to
// receive, dispatch, and acknowledge asynchronous event notifications.
type handler struct {
	logger     *zap.Logger
	dispatcher sdkservices.Dispatcher
	conns      sdkservices.Connections
	projs      sdkservices.Projects
	vars       sdkservices.Vars
}

// ns can be either:
// - "project": means project "project" with env "default".
// - "project.env": means project "project" with env "env".
func routePrefix(ns string) string {
	return fmt.Sprintf("/http/%s/", ns)
}

func Start(l *zap.Logger, mux *http.ServeMux, vs sdkservices.Vars, d sdkservices.Dispatcher, c sdkservices.Connections, p sdkservices.Projects) {
	h := handler{logger: l, dispatcher: d, conns: c, vars: vs, projs: p}
	mux.Handle(routePrefix("{ns}")+"*", h)

	// Save new autokitteh connections with user-submitted HTTP secrets.
	mux.Handle(uiPath, http.FileServer(http.FS(static.HTTPWebContent)))
	mux.HandleFunc(savePath, h.handleAuth)
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("url", r.URL.String()))

	l.Info("incoming request")

	ns := r.PathValue("ns")
	env := strings.ReplaceAll(ns, ".", "/")
	prefix := routePrefix(ns)

	url := *r.URL

	if strings.HasPrefix(url.Path, prefix) {
		url.Path = "/" + strings.TrimPrefix(url.Path, prefix)
	}

	if strings.HasPrefix(url.RawPath, prefix) {
		url.RawPath = "/" + strings.TrimPrefix(url.RawPath, prefix)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error("body read error", zap.Error(err))
		// no return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	_ = r.ParseForm()

	data := map[string]sdktypes.Value{
		"url": kittehs.Must1(sdktypes.NewStructValue(
			sdktypes.NewStringValue("url"),
			map[string]sdktypes.Value{
				"scheme":       sdktypes.NewStringValue(url.Scheme),
				"opaque":       sdktypes.NewStringValue(url.Opaque),
				"host":         sdktypes.NewStringValue(url.Host),
				"fragment":     sdktypes.NewStringValue(url.Fragment),
				"raw_fragment": sdktypes.NewStringValue(url.RawFragment),
				"raw":          sdktypes.NewStringValue(url.RawPath),
				"path":         sdktypes.NewStringValue(url.Path),
				"raw_query":    sdktypes.NewStringValue(url.RawQuery),
				"query": sdktypes.NewDictValueFromStringMap(
					kittehs.TransformMapValues(url.Query(), func(vs []string) sdktypes.Value {
						return sdktypes.NewStringValue(strings.Join(vs, ","))
					}),
				),
			},
		)),
		"method": sdktypes.NewStringValue(r.Method),
		"headers": sdktypes.NewDictValueFromStringMap(
			kittehs.TransformMapValues(r.Header, func(vs []string) sdktypes.Value {
				return sdktypes.NewStringValue(strings.Join(vs, ","))
			}),
		),
		"body": bodyToStruct(body, r.Form),
	}

	event, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: strings.ToLower(r.Method),
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("create event error", zap.Error(err))
		return
	}

	ctx := r.Context()

	pname, err := sdktypes.StrictParseSymbol(strings.SplitN(env, "/", 2)[0])
	if err != nil {
		l.Error("parse project name error", zap.Error(err))
		return
	}

	p, err := h.projs.GetByName(ctx, pname)
	if err != nil {
		l.Error("get project error", zap.Error(err))
		return
	}

	conns, err := h.conns.List(ctx, sdkservices.ListConnectionsFilter{
		IntegrationID: IntegrationID,
		ProjectID:     p.ID(),
	})
	if err != nil {
		l.Error("list connections error", zap.Error(err))
		return
	}

	for _, c := range conns {
		cid := c.ID()
		l := l.With(zap.String("cid", cid.String()))
		eid, err := h.dispatcher.Dispatch(ctx, event.WithConnectionID(cid), &sdkservices.DispatchOptions{Env: env})
		if err != nil {
			l.Error("dispatch error", zap.Error(err))
		}

		l.Info("dispatched", zap.String("event_id", eid.String()))
	}
}
