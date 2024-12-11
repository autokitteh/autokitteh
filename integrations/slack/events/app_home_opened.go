package events

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type AppHomeOpen struct {
	Type    string `json:"type,omitempty"`
	User    string `json:"user,omitempty"`
	Channel string `json:"channel,omitempty"`
	Tab     string `json:"tab,omitempty"`
	EventTS string `json:"event_ts,omitempty"`
}

type appHomeOpenContainer struct {
	Event *AppHomeOpen `json:"event"`
}

func AppHomeOpenHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
	// Ignore self-triggered events.
	for _, a := range cb.Authorizations {
		if a.UserID == cb.Event.User {
			l.Debug("Ignoring self-triggered event")
			return nil
		}
	}

	var j appHomeOpenContainer
	if err := json.Unmarshal(body, &j); err != nil {
		invalidEventError(l, w, body, err)
		return nil
	}

	return j.Event
}
