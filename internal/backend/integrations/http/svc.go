package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	IntegrationName = sdktypes.NewSymbol("http")
	IntegrationID   = fixtures.NewBuiltinIntegrationID("http")
	IntegrationDesc = sdktypes.NewIntegration(IntegrationID, IntegrationName)
	Integration     = sdkintegrations.NewIntegration(IntegrationDesc, nil)
)

type svc struct {
	l          *zap.Logger
	dispatcher sdkservices.Dispatcher
	conns      sdkservices.Connections
	projs      sdkservices.Projects
}

// ns can be either:
// - "project": means project "project" with env "default".
// - "project.env": means project "project" with env "env".
func routePrefix(ns string) string {
	return fmt.Sprintf("/http/%s/", ns)
}

func Start(l *zap.Logger, mux *http.ServeMux, d sdkservices.Dispatcher, c sdkservices.Connections, p sdkservices.Projects) {
	s := svc{l: l, dispatcher: d, conns: c, projs: p}
	mux.Handle(routePrefix("{ns}")+"*", &s)
}

func (s *svc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := s.l.With(zap.String("urlPath", r.URL.Path))
	l.Info("Incoming HTTP request")

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
		l.Error("Failed to convert protocol buffer to SDK event",
			zap.String("eventType", strings.ToLower(r.Method)),
			zap.Any("data", data),
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	pname, err := sdktypes.StrictParseSymbol(strings.SplitN(env, "/", 2)[0])
	if err != nil {
		l.Debug("parse project name error", zap.Error(err))
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	p, err := s.projs.GetByName(ctx, pname)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			l.Debug("project not found", zap.String("project", pname.String()))
		} else {
			l.Error("get project", zap.Error(err))
		}
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Retrieve all the relevant connections for this event.
	conns, err := s.conns.List(ctx, sdkservices.ListConnectionsFilter{
		IntegrationID: IntegrationID,
		ProjectID:     p.ID(),
	})
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	s.dispatch(ctx, conns, env, event)
}

func (s *svc) dispatch(ctx context.Context, cs []sdktypes.Connection, env string, event sdktypes.Event) {
	for _, c := range cs {
		cid := c.ID()
		opts := &sdkservices.DispatchOptions{Env: env}
		eid, err := s.dispatcher.Dispatch(ctx, event.WithConnectionID(cid), opts)
		l := s.l.With(
			zap.String("connectionID", cid.String()),
			zap.String("eventID", eid.String()),
		)
		if err != nil {
			l.Error("Event dispatch failed", zap.Error(err))
			return
		}
		l.Debug("Event dispatched")
	}
}

func bodyToStruct(body []byte, form url.Values) sdktypes.Value {
	var (
		v        any
		jsonBody sdktypes.Value
	)

	jsonDecoder := json.NewDecoder(bytes.NewReader(body))
	jsonDecoder.UseNumber()

	if err := jsonDecoder.Decode(&v); err != nil {
		jsonBody = kittehs.Must1(sdktypes.NewConstFunctionError("json", err))
	} else if vv, err := sdktypes.WrapValue(v); err != nil {
		jsonBody = kittehs.Must1(sdktypes.NewConstFunctionError("json", err))
	} else {
		jsonBody = kittehs.Must1(sdktypes.NewConstFunctionValue("json", vv))
	}

	// add form() only for requests (when not nil) and not for responses
	if form != nil {
		formBody := kittehs.Must1(sdktypes.NewConstFunctionValue("form", sdktypes.NewDictValueFromStringMap(
			kittehs.TransformMapValues(form, func(vs []string) sdktypes.Value {
				return sdktypes.NewStringValue(strings.Join(vs, ","))
			}),
		)))
		return kittehs.Must1(sdktypes.NewStructValue(
			sdktypes.NewStringValue("body"),
			map[string]sdktypes.Value{
				"text":  kittehs.Must1(sdktypes.NewConstFunctionValue("text", sdktypes.NewStringValue(string(body)))),
				"bytes": kittehs.Must1(sdktypes.NewConstFunctionValue("bytes", sdktypes.NewBytesValue(body))),
				"json":  jsonBody,
				"form":  formBody,
			},
		))
	}
	return kittehs.Must1(sdktypes.NewStructValue(
		sdktypes.NewStringValue("body"),
		map[string]sdktypes.Value{
			"text":  kittehs.Must1(sdktypes.NewConstFunctionValue("text", sdktypes.NewStringValue(string(body)))),
			"bytes": kittehs.Must1(sdktypes.NewConstFunctionValue("bytes", sdktypes.NewBytesValue(body))),
			"json":  jsonBody,
		},
	))
}
