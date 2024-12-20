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

	"go.jetify.com/typeid"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const WebhooksPathPrefix = "/webhooks/"

type Service struct {
	logger   *zap.Logger
	dispatch sdkservices.DispatchFunc
	db       db.DB
}

func New(l *zap.Logger, db db.DB, dispatch sdkservices.DispatchFunc) *Service {
	return &Service{logger: l, db: db, dispatch: dispatch}
}

func (s *Service) Start(muxes *muxes.Muxes) {
	muxes.NoAuth.Handle(WebhooksPathPrefix+"{slug}", s)
	muxes.NoAuth.Handle(WebhooksPathPrefix+"{slug}/", s)
}

func InitTrigger(trigger sdktypes.Trigger) sdktypes.Trigger {
	unique := typeid.Must(typeid.FromUUIDWithPrefix("", sdktypes.NewUUID().String()))
	return trigger.WithWebhookSlug(unique.String())
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	sl := s.logger.With(
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("slug", slug),
	).Sugar()

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
		Memo: map[string]string{
			"method":       r.Method,
			"webhook_slug": slug,
			"remote_addr":  r.RemoteAddr,
			"trigger_id":   t.ID().String(),
			"trigger_uuid": t.ID().UUIDValue().String(),
		},
	})
	if err != nil {
		sl.Errorw("failed to convert protocol buffer to event", "event_type", r.Method, "data", data, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	eid, err := s.dispatch(authcontext.SetAuthnSystemUser(ctx), event, nil)
	if err != nil {
		sl.Errorw("dispatch failed", "event", event, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("AutoKitteh-Event-ID", eid.String())
	w.WriteHeader(http.StatusAccepted)
}

func requestToData(r *http.Request) (map[string]sdktypes.Value, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("body read: %w", err)
	}

	r.Body = io.NopCloser(bytes.NewReader(body))
	_ = r.ParseForm()

	return map[string]sdktypes.Value{
		"body": bodyData(body, r.PostForm),
		"headers": sdktypes.NewDictValueFromStringMap(
			kittehs.TransformMapValues(r.Header, func(vs []string) sdktypes.Value {
				return sdktypes.NewStringValue(strings.Join(vs, ", "))
			}),
		),
		"trailers": sdktypes.NewDictValueFromStringMap(
			kittehs.TransformMapValues(r.Trailer, func(vs []string) sdktypes.Value {
				return sdktypes.NewStringValue(strings.Join(vs, ", "))
			})),
		"method":  sdktypes.NewStringValue(r.Method),
		"raw_url": sdktypes.NewStringValue(r.RequestURI),
		"url":     urlData(r.URL),
	}, nil
}

func bodyData(body []byte, form url.Values) sdktypes.Value {
	bytes := sdktypes.Nothing
	if len(body) > 0 {
		bytes = sdktypes.NewBytesValue(body)
	}

	return sdktypes.NewDictValueFromStringMap(
		map[string]sdktypes.Value{
			"bytes": bytes,
			"form":  formData(form),
			"json":  jsonData(body),
		},
	)
}

func formData(form url.Values) sdktypes.Value {
	if len(form) == 0 {
		return sdktypes.Nothing
	}

	return sdktypes.NewDictValueFromStringMap(
		kittehs.TransformMapValues(form, func(vs []string) sdktypes.Value {
			return sdktypes.NewStringValue(strings.Join(vs, ", "))
		}),
	)
}

func jsonData(body []byte) sdktypes.Value {
	var v any
	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()

	if err := d.Decode(&v); err != nil {
		return sdktypes.Nothing
	} else if vv, err := sdktypes.WrapValue(v); err != nil {
		return sdktypes.Nothing
	} else {
		return vv
	}
}

func urlData(u *url.URL) sdktypes.Value {
	return sdktypes.NewDictValueFromStringMap(
		map[string]sdktypes.Value{
			"fragment": sdktypes.NewStringValue(u.Fragment),
			"path":     sdktypes.NewStringValue(u.Path),
			"query": sdktypes.NewDictValueFromStringMap(
				kittehs.TransformMapValues(u.Query(), func(vs []string) sdktypes.Value {
					return sdktypes.NewStringValue(strings.Join(vs, ", "))
				}),
			),
			"raw_fragment": sdktypes.NewStringValue(u.RawFragment),
			"raw_path":     sdktypes.NewStringValue(u.RawPath),
			"raw_query":    sdktypes.NewStringValue(u.RawQuery),
		},
	)
}
