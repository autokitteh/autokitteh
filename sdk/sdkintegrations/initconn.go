package sdkintegrations

import (
	"fmt"
	"net/http"
	"net/url"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// FinalizeConnectionInit finalizes all integration-specific connection flows.
// It encodes their resulting details and redirects the user to the final HTTP
// handler ("post-init") that saves the data in the connection's scope, and
// redirects the user to the last HTTP response, based on its origin
// (local AK server / local VS Code extension / SaaS web UI).
func FinalizeConnectionInit(w http.ResponseWriter, r *http.Request, iid sdktypes.IntegrationID, data []sdktypes.Var) {
	vars, err := kittehs.EncodeURLData(data)
	if err != nil {
		http.Error(w, "Failed to encode URL vars", http.StatusInternalServerError)
		return
	}

	cid, err := sdktypes.ParseConnectionID(r.URL.Query().Get("cid"))
	if err != nil {
		http.Error(w, "Failed to parse connection ID", http.StatusBadRequest)
		return
	}

	id := cid.String()
	if id == "" {
		// The user needs to select a specific connection at the end,
		// based on the integration, if it wasn't selected at the beginning.
		id = iid.String()
	}

	origin := url.QueryEscape(r.URL.Query().Get("origin"))

	u := fmt.Sprintf("/connections/%s/postinit?vars=%s&origin=%s", id, vars, origin)
	http.Redirect(w, r, u, http.StatusFound)
}
