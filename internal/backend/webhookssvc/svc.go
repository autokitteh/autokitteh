package webhookssvc

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
	"time"

	"go.jetify.com/typeid"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// An unwrapper that is always safe to serialize to string afterwards.
var unwrapper = sdktypes.ValueWrapper{
	SafeForJSON:         true,
	UnwrapStructsAsJSON: true,
}

const WebhooksPathPrefix = "/webhooks/"

type Config struct {
	SessionOutcomePollInterval time.Duration `koanf:"session_outcome_poll_interval"`
	WebhookResponseTimeout     time.Duration `koanf:"webhook_response_timeout"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		WebhookResponseTimeout:     30 * time.Second,
		SessionOutcomePollInterval: 100 * time.Millisecond,
	},
}

type Service struct {
	logger   *zap.Logger
	dispatch sdkservices.DispatchFunc
	db       db.DB
	cfg      *Config
}

func New(l *zap.Logger, cfg *Config, db db.DB, dispatch sdkservices.DispatchFunc) *Service {
	return &Service{logger: l, db: db, dispatch: dispatch, cfg: cfg}
}

func (s *Service) Start(muxes *muxes.Muxes) {
	muxes.NoAuth.Handle(WebhooksPathPrefix+"{slug}", s)
	muxes.NoAuth.Handle(WebhooksPathPrefix+"{slug}/", s)
}

func InitTrigger(trigger sdktypes.Trigger) sdktypes.Trigger {
	unique := typeid.Must(typeid.FromUUIDWithPrefix("", sdktypes.NewUUID().String()))
	return trigger.WithWebhookSlug(unique.String())
}

func WebhookSlugToAddress(slug string) (string, error) {
	return url.JoinPath(fixtures.ServiceBaseURL(), WebhooksPathPrefix, slug)
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

	t, err := s.db.GetTriggerWithActiveDeploymentByWebhookSlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			sl.Infof("could not find an active deployment for trigger by slug %q", slug)
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		sl.Errorw("get trigger by slug failed", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	sl.With("trigger", t).Infof("webhook request: method=%s, trigger_event_type=%s", r.Method, t.EventType())

	data, err := requestToData(r, slug)
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

	isSync := t.IsSync()

	sl = sl.With("sync", isSync, "event_id", event.ID())

	opts := &sdkservices.DispatchOptions{Wait: isSync}

	resp, err := s.dispatch(authcontext.SetAuthnSystemUser(ctx), event, opts)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrResourceExhausted) {
			sl.Warnw("dispatch failed: resource exhausted", "event", event, "err", err)
			http.Error(w, "Resource Exhausted", http.StatusTooManyRequests)
			return
		}
		sl.Errorw("dispatch failed", "event", event, "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("AutoKitteh-Event-ID", resp.EventID.String())
	w.Header().Set("Cache-Control", "no-cache")

	if !isSync {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	s.handleSyncResponse(ctx, w, resp.SessionIDs, sl)
}

func (s *Service) handleSyncResponse(ctx context.Context, w http.ResponseWriter, sids []sdktypes.SessionID, sl *zap.SugaredLogger) {
	w.Header().Set("Connection", "keep-alive")

	if len(sids) == 0 {
		sl.Warnw("no session was created for event")
		http.Error(w, "No session was created for event", http.StatusBadGateway)
		return
	}

	if len(sids) > 1 {
		sl.Warnw("multiple sessions were created for event", "session_ids", sids)
		http.Error(w, "Multiple sessions were created for event", http.StatusBadGateway)
		return
	}

	session, err := s.db.GetSession(ctx, sids[0])
	if err != nil {
		sl.Errorw("get session failed", "session_id", sids[0], "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("AutoKitteh-Session-ID", session.ID().String())

	if tmo := s.cfg.WebhookResponseTimeout; tmo != 0 {
		var done context.CancelFunc
		ctx, done = context.WithTimeout(ctx, tmo)
		defer done()
	}

	var (
		firstOutcomeHandled bool
		skip                int32 // no need to reprocess records we've already seen.
	)

	for {
		if skip > 0 {
			// delay poll only after first iteration.

			select {
			case <-time.After(s.cfg.SessionOutcomePollInterval):
				// nop
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					// timeout
					sl.Debugw("timed out waiting for session to complete")
					w.WriteHeader(http.StatusGatewayTimeout)
				} else {
					// cancelled
					sl.Debugw("request context cancelled", "err", ctx.Err())
					w.WriteHeader(http.StatusRequestTimeout)
				}
				return
			}
		}

		log, err := s.db.GetSessionLog(
			ctx,
			sdkservices.SessionLogRecordsFilter{
				SessionID: session.ID(),
				Types:     sdktypes.OutcomeSessionLogRecordType | sdktypes.StateSessionLogRecordType,
				PaginationRequest: sdktypes.PaginationRequest{
					Ascending: true,
					Skip:      skip,
				},
			},
		)
		if err != nil {
			sl.Errorw("get session log failed", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for _, r := range log.Records {
			skip++

			if state := r.GetState(); state.IsValid() {
				switch state.Type() {
				case sdktypes.SessionStateTypeError:
					sl.Warnw("session ended in error", "state", state)
					w.WriteHeader(http.StatusBadGateway)

				case sdktypes.SessionStateTypeStopped:
					sl.Warnw("session was stopped before producing an outcome", "state", state)
					w.WriteHeader(http.StatusBadGateway)

				case sdktypes.SessionStateTypeCompleted:
					w.WriteHeader(http.StatusOK)

				default:
					// not final
					continue
				}

				return
			}

			if v := r.GetOutcome(); v.IsValid() {
				outcome, err := parseOutcomeValue(v)
				if err != nil {
					sl.Errorw("failed to parse outcome value", "err", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				more, err := s.writeOutcome(w, outcome, firstOutcomeHandled)
				if err != nil {
					sl.Errorw("failed to handle outcome", "err", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				if !more {
					return
				}

				firstOutcomeHandled = true
			}
		}
	}
}

func (s *Service) writeOutcome(w http.ResponseWriter, outcome httpOutcome, firstOutcomeHandled bool) (more bool, err error) {
	if !firstOutcomeHandled {
		var hadContentType bool

		for k, v := range outcome.Headers {
			w.Header().Set(k, v)

			if strings.ToLower(k) == "content-type" {
				hadContentType = true
			}
		}

		if !hadContentType && outcome.Json.IsValid() {
			w.Header().Set("Content-Type", "application/json")
		}

		code := outcome.StatusCode
		if code == 0 {
			code = http.StatusOK
		}

		w.WriteHeader(code)
	}

	if err := outcome.WriteBody(w); err != nil {
		return false, fmt.Errorf("write outcome body: %w", err)
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return outcome.More, nil
}

type httpOutcome struct {
	StatusCode int
	Body       sdktypes.Value
	Json       sdktypes.Value // due to the way unwrapping work, this must be "Json" and not "JSON".
	Headers    map[string]string
	More       bool
}

func parseOutcomeValue(v sdktypes.Value) (outcome httpOutcome, err error) {
	err = sdktypes.UnwrapValueInto(&outcome, v)
	return
}

// WriteBody writes body bytes to a writer.
//
// If both Body and JSON are set, an error is returned.
//
// If Body is set, it is converted to bytes as follows:
//   - If Body is a string, it is converted to []byte of the string.
//   - If Body is []byte, it is returned as-is.
//   - Otherwise, Body is marshaled to JSON and the resulting bytes are returned.
//
// If JSON is set, it is marshaled to JSON and the resulting bytes are returned.
//
// If neither Body nor JSON is set, nil and no error are returned.
func (o httpOutcome) WriteBody(w io.Writer) error {
	var b sdktypes.Value

	switch {
	case o.Body.IsValid() && o.Json.IsValid():
		return errors.New("outcome cannot have both 'body' and 'json' fields set together")
	case o.Body.IsValid():
		if v := o.Body.GetString(); v.IsValid() {
			if _, err := w.Write([]byte(v.Value())); err != nil {
				return fmt.Errorf("write body string: %w", err)
			}
			return nil
		}

		if v := o.Body.GetBytes(); v.IsValid() {
			if _, err := w.Write(v.Value()); err != nil {
				return fmt.Errorf("write body bytes: %w", err)
			}
			return nil
		}

		b = o.Body
	case o.Json.IsValid():
		b = o.Json
	default:
		// nothing to write
		return nil
	}

	u, err := unwrapper.Unwrap(b)
	if err != nil {
		return fmt.Errorf("outcome body unwrap: %w", err)
	}

	if err := json.NewEncoder(w).Encode(u); err != nil {
		return fmt.Errorf("outcome json marshal: %w", err)
	}

	return nil
}

func requestToData(r *http.Request, slug string) (map[string]sdktypes.Value, error) {
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
		"url":     urlData(r.URL, slug),
	}, nil
}

func bodyData(body []byte, form url.Values) sdktypes.Value {
	text, bytes := sdktypes.Nothing, sdktypes.Nothing

	if len(body) > 0 {
		bytes = sdktypes.NewBytesValue(body)
		text = sdktypes.NewStringValue(string(body))
	}

	return sdktypes.NewDictValueFromStringMap(
		map[string]sdktypes.Value{
			"bytes": bytes,
			"text":  text,
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

func urlData(u *url.URL, slug string) sdktypes.Value {
	pathSuffix := strings.TrimPrefix(u.Path, WebhooksPathPrefix+slug)

	return sdktypes.NewDictValueFromStringMap(
		map[string]sdktypes.Value{
			"path": sdktypes.NewStringValue(u.Path),
			"query": sdktypes.NewDictValueFromStringMap(
				kittehs.TransformMapValues(u.Query(), func(vs []string) sdktypes.Value {
					return sdktypes.NewStringValue(strings.Join(vs, ", "))
				}),
			),
			"path_suffix": sdktypes.NewStringValue(pathSuffix),
		},
	)
}
