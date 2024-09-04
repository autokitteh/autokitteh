package webhookssvc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lithammer/shortuuid/v4"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const WebhooksPathPrefix = "/webhooks/"

type Service struct {
	sl         *zap.SugaredLogger
	dispatcher sdkservices.Dispatcher
	db         db.DB
}

func New(l *zap.Logger, db db.DB, dispatcher sdkservices.Dispatcher) *Service {
	return &Service{sl: l.Sugar(), db: db, dispatcher: dispatcher}
}

func (s *Service) Start(muxes *muxes.Muxes) {
	muxes.NoAuth.Handle(WebhooksPathPrefix+"{slug}", s)
	muxes.NoAuth.Handle(WebhooksPathPrefix+"{slug}/*", s)
}

func InitTrigger(trigger sdktypes.Trigger) sdktypes.Trigger {
	return trigger.WithWebhookSlug(shortuuid.DefaultEncoder.Encode(sdktypes.NewUUID()))
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	sl := s.sl.With("url", r.URL.String(), "method", r.Method, "slug", slug)
	sl.Infof("webhook request: %s %s", r.Method, r.URL.Path)

	ctx := r.Context()

	t, err := s.db.GetTriggerByWebhookSlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			sl.Infof("slug %q not found", slug)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		sl.Errorw("get trigger by slug failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data, err := requestToData(r)
	if err != nil {
		sl.Errorw("failed to convert request to data", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	event, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType:     strings.ToLower(r.Method),
		Data:          kittehs.TransformMapValues(data, sdktypes.ToProto),
		DestinationId: t.ID().String(),
	})
	if err != nil {
		sl.Errorw("failed to convert protocol buffer to event", "event_type", r.Method, "data", data, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if _, err := s.dispatcher.Dispatch(ctx, event, nil); err != nil {
		sl.Errorw("dispatch failed", "event", event, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func requestToData(r *http.Request) (map[string]sdktypes.Value, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("body read: %w", err)
	}

	r.Body = io.NopCloser(bytes.NewReader(body))
	_ = r.ParseForm()

	url := r.URL

	return map[string]sdktypes.Value{
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
	}, nil
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
