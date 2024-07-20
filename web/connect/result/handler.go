package result

import (
	"html/template"
	"net/http"
	"strconv"

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
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request, t *template.Template, ok bool) {
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

	// Read the intended error status code from the URL query.
	statusCode := http.StatusOK
	if !ok {
		status := r.URL.Query().Get("status")
		if status == "" {
			status = "400"
		}

		statusCode, err = strconv.Atoi(status)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}

	// Generate the HTML page with the integration details.
	i := integ.Get()
	data := map[string]string{
		"cid":    cid.String(),
		"origin": r.URL.Query().Get("origin"),
		"error":  r.URL.Query().Get("error"),
		"integ":  i.UniqueName().String(),
		"name":   i.DisplayName(),
		"logo":   i.LogoURL().String(),
	}

	w.WriteHeader(statusCode)

	if err := t.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
