package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lithammer/shortuuid/v4"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	IntegrationID = fixtures.NewBuiltinIntegrationID("webhooks")

	desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
		IntegrationId: IntegrationID.String(),
		UniqueName:    "webhooks",
	}))

	slugVarName = sdktypes.NewSymbol("slug")
)

type Service struct {
	sl         *zap.SugaredLogger
	dispatcher sdkservices.Dispatcher
	vars       sdkservices.Vars
}

func webhookPath(slug string) string { return "/webhooks/" + slug }

func New(l *zap.Logger, vars sdkservices.Vars) (sdkservices.Integration, *Service) {
	s := Service{
		sl:   l.Sugar(),
		vars: vars,
	}

	return sdkintegrations.NewIntegration(
		desc,
		nil,
		sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
			slug, err := getSlug(ctx, vars, cid)
			if err != nil {
				return sdktypes.Status{}, fmt.Errorf("get slug: %w", err)
			}

			if slug == "" {
				return sdktypes.NewStatus(sdktypes.StatusCodeError, "slug not set"), nil
			}

			return sdktypes.NewStatus(sdktypes.StatusCodeOK, webhookPath(slug)), nil
		}),
	), &s
}

func (s *Service) Start(muxes *muxes.Muxes, d sdkservices.Dispatcher) {
	s.dispatcher = d
	muxes.NoAuth.Handle(webhookPath("{slug}")+"/*", s)
}

func getSlug(ctx context.Context, vars sdkservices.Vars, cid sdktypes.ConnectionID) (string, error) {
	vs, err := vars.Get(ctx, sdktypes.NewVarScopeID(cid), slugVarName)
	if err != nil {
		return "", fmt.Errorf("get vars(%v): %w", cid, err)
	}

	if len(vs) > 0 {
		return vs[0].Value(), nil
	}

	return "", nil
}

// Called from connections service after a Webhooks connection was created.
// A Webhook connection needs to be customized by the integration upon creation: it needs a unique
// endpoing slug. This also tells the connections to update the connection record with the new
// slug as a link.
func (s *Service) ConnectionCreated(ctx context.Context, conn sdktypes.Connection) (sdktypes.Connection, error) {
	sl := s.sl.With("connection_id", conn.ID())

	slug := shortuuid.New()
	link := webhookPath(slug)

	sl.With("slug", slug).Infof("created webhook %q for connection %v", link, conn.ID())

	// Set the slug as a connection variable, so that the webhook handler can look it up.
	if err := s.vars.Set(ctx, sdktypes.NewVar(slugVarName).SetValue(slug).WithScopeID(sdktypes.NewVarScopeID(conn.ID()))); err != nil {
		return sdktypes.InvalidConnection, fmt.Errorf("set vars(%v): %w", conn.ID(), err)
	}

	return conn.WithLink("webhook_url", link), nil
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	sl := s.sl.With("url", r.URL.String(), "method", r.Method, "slug", slug)
	sl.Infof("webhook request: %s %s", r.Method, r.URL.Path)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sl.Error("body read error", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusBadGateway)
		return
	}

	r.Body = io.NopCloser(bytes.NewReader(body))
	_ = r.ParseForm()

	url := r.URL

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
				"slug": sdktypes.NewStringValue(slug),
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
		sl.Errorw("failed to convert protocol buffer to event", "event_type", r.Method, "data", data, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Look up connections based on the slug in the path.
	cids, err := s.vars.FindConnectionIDs(ctx, IntegrationID, slugVarName, slug)
	if err != nil {
		sl.Errorw("failed to find connection ids", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	s.dispatch(ctx, cids, event)
}

func (s *Service) dispatch(ctx context.Context, cids []sdktypes.ConnectionID, event sdktypes.Event) {
	for _, cid := range cids {
		sl := s.sl.With("connection_id", cid)

		_, err := s.dispatcher.Dispatch(ctx, event.WithConnectionID(cid), nil)
		if err != nil {
			sl.Errorw("Event dispatch failed", "err", err)
			return
		}
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
