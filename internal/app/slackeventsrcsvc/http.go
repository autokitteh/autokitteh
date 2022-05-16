package slackeventsrcsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func (s *Svc) initHTTP(r *mux.Router) {
	if s.Config.OAuthEnabled {
		if r == nil {
			panic("router required")
		}

		r.HandleFunc("/slack/oauth/install", s.httpOAuthInstall).Methods("GET")
	}

	if r != nil {
		r.HandleFunc("/slack/event", s.httpEvent).Methods("POST")
	}
}

func (s *Svc) httpEvent(w http.ResponseWriter, r *http.Request) {
	l := s.L

	var buf bytes.Buffer

	if _, err := buf.ReadFrom(r.Body); err != nil {
		l.Error("read failed", "err", err)
		http.Error(w, fmt.Sprintf("body read error: %v", err), http.StatusBadGateway)
		return
	}

	body := buf.String()

	v, err := slack.NewSecretsVerifier(r.Header, s.Config.SigningSecret)
	if err != nil {
		l.Error("NewSecretsVerifier error", "header", r.Header, "err", err)
		http.Error(w, "secret verification error", http.StatusBadRequest)
		return
	}

	if _, err := v.Write(buf.Bytes()); err != nil {
		l.Error("verifier write error", "err", err)
		http.Error(w, "verifier write error", http.StatusBadRequest)
		return
	}

	if err := v.Ensure(); err != nil {
		l.Warn("verifier ensure error", "err", err)
		http.Error(w, "verification error", http.StatusBadRequest)
		return
	}

	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		l.Warn("event parsing failed", "err", err)
		http.Error(w, "parse error", http.StatusBadRequest)
		return
	}

	l.Debug("event received", "event", event)

	switch event.Type {
	case slackevents.URLVerification:
		e := event.Data.(*slackevents.EventsAPIURLVerificationEvent)

		w.Header().Set("Content-Type", "text")
		_, _ = w.Write([]byte(e.Challenge))

		l.Debug("responded to url verification")
		return

	case slackevents.CallbackEvent:
		data := event.Data.(*slackevents.EventsAPICallbackEvent)

		s.handleCallbackEvent(context.Background(), data, &event.InnerEvent)
	}
}
