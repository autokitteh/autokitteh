package microsoft

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// https://learn.microsoft.com/en-us/graph/api/resources/changenotificationcollection
type changeNotifs struct {
	ValidationTokens []string         `json:"validationTokens"`
	Value            []map[string]any `json:"value"`
}

// handleEvent receives asynchronous events ("change notifications") from Microsoft
// Graph, and dispatches them to zero or more AutoKitteh connections. Subscriptions
// are created and renewed automatically by AutoKitteh - see "subscriptions.go" and
// "lifecycle.go", respectively, for implementation details. See also:
// https://learn.microsoft.com/en-us/graph/change-notifications-overview
// https://learn.microsoft.com/en-us/graph/api/resources/change-notifications-api-overview
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks
// https://learn.microsoft.com/en-us/graph/change-notifications-with-resource-data
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("url_path", r.URL.Path))
	defer r.Body.Close()

	// Validate new subscriptions.
	if vt := r.FormValue("validationToken"); vt != "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(vt)); err != nil {
			l.Warn("MS change notif: failed to write validation token", zap.Error(err))
		}
		return
	}

	// Parse incoming events.
	var notifs changeNotifs
	if err := json.NewDecoder(r.Body).Decode(&notifs); err != nil {
		h.logger.Warn("MS change notif: failed to parse request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// TODO(INT-203): Dispatch change notifications to AutoKitteh connections.
}
