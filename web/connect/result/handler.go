package result

import (
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type handler struct {
	connections  sdkservices.Connections
	integrations sdkservices.Integrations
}

func NewHandler(s sdkservices.Services) handler {
	return handler{connections: s.Connections(), integrations: s.Integrations()}
}

// ServeHTTP is an HTTP handler that serves generic
// connection success and error webpages for all integrations.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request, t *template.Template) {
	// Read the connection ID from the request path.
	cid, err := sdktypes.StrictParseConnectionID(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Bad connection ID", http.StatusBadRequest)
		return
	}

	// Fetch the connection details.
	conn, err := sdkerrors.IgnoreNotFoundErr(h.connections.Get(r.Context(), cid))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !conn.IsValid() {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	// Fetch the connection's integration details.
	integ, err := h.integrations.GetByID(r.Context(), conn.IntegrationID())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	i := integ.Get()
	data := map[string]string{
		"id":   i.UniqueName().String(),
		"name": i.DisplayName(),
		"logo": i.LogoURL().String(),
	}

	// Generate the HTML page with the integration details.
	if err := t.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
