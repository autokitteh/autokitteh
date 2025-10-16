package webhooks

import (
	"encoding/json"
	"net/http"

	"github.com/infracloudio/msbotbuilder-go/schema"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// MessagePath is the URL path for our webhook to handle message events.
	MessagePath = "/azurebot/message"
)

// HandleMessage dispatches to autokitteh an asynchronous event notification.
func (h handler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	var act schema.Activity
	if err := json.NewDecoder(r.Body).Decode(&act); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tenantID := act.Conversation.TenantID

	l := h.logger.With(zap.String("tenant_id", tenantID))

	l.Debug("received activity for tenant "+tenantID, zap.Any("activity", act))

	cids, err := h.vars.FindActiveConnectionIDs(r.Context(), h.integrationID, sdktypes.NewSymbol("tenant_id"), tenantID)
	if err != nil {
		l.Warn("failed to find connection IDs for tenant "+tenantID, zap.Error(err))
		http.Error(w, "no connections configured for tenant", http.StatusBadRequest)
		return

	}

	v, err := sdktypes.WrapValue(act)
	if err != nil {
		l.Error("failed to wrap activity for tenant"+tenantID, zap.Error(err))
		http.Error(w, "failed to wrap activity", http.StatusInternalServerError)
		return
	}

	m, err := v.ToStringValuesMap()
	if err != nil {
		l.Error("failed to convert wrapped event for tenant "+tenantID, zap.Error(err))
		http.Error(w, "failed to convert wrapped event", http.StatusInternalServerError)
		return
	}

	evt, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: string(act.Type),
		Data:      kittehs.TransformMapValues(m, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("failed to convert protocol buffer to SDK event for tenant "+tenantID, zap.Error(err))
		http.Error(w, "failed to convert protocol buffer to SDK event", http.StatusInternalServerError)
		return
	}

	common.DispatchEvent(r.Context(), h.logger, h.dispatch, evt, cids)

	w.WriteHeader(http.StatusOK)
}
