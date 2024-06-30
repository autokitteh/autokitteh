package sdkintegrations

import (
	"fmt"
	"net/http"
	"net/url"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// FinalizeConnectionInit finalizes the connection initialization.
// This is done by encoding the init data into the connection initialization URL
// and redirecting the user there. The destination handler will save `data` in
// the connection's scope.
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
		id = iid.String()
	}

	method := url.QueryEscape(r.URL.Query().Get("method"))

	u := fmt.Sprintf("/connections/%s/postinit?vars=%s&method=%s", id, vars, method)
	http.Redirect(w, r, u, http.StatusSeeOther)
}
