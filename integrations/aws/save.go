package aws

import (
	"fmt"
	"net/http"
	"strings"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	HeaderContentType = "Content-Type"
	ContentTypeForm   = "application/x-www-form-urlencoded"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to save data from web form submissions as connections.
type handler struct {
	vars sdkservices.Vars
}

func NewHTTPHandler(vars sdkservices.Vars) http.Handler {
	return handler{vars: vars}
}

// ServeHTTP saves a new autokitteh connection with user-submitted data.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check "Content-Type" header.
	ct := r.Header.Get(HeaderContentType)
	if ct != ContentTypeForm {
		http.Error(w, fmt.Sprintf("Unexpected Content-Type header: %q", ct), http.StatusBadRequest)
		return
	}

	// Read and parse POST request body.
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Form parsing error: "+err.Error(), http.StatusBadRequest)
		return
	}

	initData := sdktypes.EncodeVars(authData{
		Region:      strings.TrimSpace(r.FormValue("region")),
		AccessKeyID: strings.TrimSpace(r.FormValue("access_key")),
		SecretKey:   strings.TrimSpace(r.FormValue("secret_key")),
		Token:       strings.TrimSpace(r.FormValue("token")),
	})

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}
