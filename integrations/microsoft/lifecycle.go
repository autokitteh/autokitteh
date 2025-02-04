package microsoft

import (
	"io"
	"net/http"

	"go.uber.org/zap"
)

// handleLifecycle receives asynchronous events ("lifecycle notifications") from
// Microsoft Graph, to renew "change notification" subscriptions automatically on
// behalf of AutoKitteh connections. In addition to "event.go" links, see also:
// https://learn.microsoft.com/en-us/graph/change-notifications-lifecycle-events
func (h handler) handleLifecycle(w http.ResponseWriter, r *http.Request) {
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

	// TODO(INT-203): Use "json.NewDecoder" instead of "io.ReadAll", like in "event.go".
	// TODO(INT-203): Consider merging this into "event.go"?
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Warn("MS lifecycle notif: failed to read request body", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	l.Warn("TODO: handle MS lifecycle notif",
		zap.String("url", r.URL.String()),
		zap.Any("headers", r.Header),
		zap.ByteString("body", body),
	)

	w.WriteHeader(http.StatusAccepted)
}
