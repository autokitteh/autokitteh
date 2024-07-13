package google

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// HandleCreds saves a new AutoKitteh connection with a user-submitted JSON key.
// It also acts as a passthrough for the OAuth connection mode, to save optional
// details (e.g. Google Form ID), to support and manage incoming events.
func (h handler) HandleCreds(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.Abort("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.Abort("form parsing error")
		return
	}

	// Validate input: Google Forms ID.
	formID := r.PostFormValue("form_id")
	if formID != "" {
		ok, err := regexp.MatchString(`[\w-]{20,}`, formID)
		if err != nil {
			l.Error("Failed to validate form ID", zap.Error(err))
			c.Abort(fmt.Sprintf("form ID validation error: %v", err))
			return
		}
		if !ok {
			c.Abort(fmt.Sprintf("invalid Google Forms ID %q", formID))
			return
		}
	}

	switch r.PostFormValue("auth_type") {
	// GCP service-account JSON-key connection? Save the JSON key.
	case "json":
		// TODO(ENG-1103): Create watches for the form's events, if the ID isn't empty.
		c.Finalize(sdktypes.EncodeVars(&vars.Vars{JSON: r.PostFormValue("json"), FormID: formID}))

	// User OAuth connect? Redirect to AutoKitteh's OAuth starting point.
	case "oauth":
		if err := h.saveFormID(r.Context(), c, formID); err != nil {
			l.Error("Connection ID parsing error", zap.Error(err))
			e := fmt.Sprintf("form ID saving error: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError, e)
			return
		}
		http.Redirect(w, r, oauthURL(c, r.PostForm), http.StatusFound)

	// Unknown mode.
	default:
		err := fmt.Sprintf("unexpected auth type %q", r.PostFormValue("auth_type"))
		c.AbortWithStatus(http.StatusInternalServerError, err)
	}
}

func (h handler) saveFormID(ctx context.Context, c sdkintegrations.ConnectionInit, formID string) error {
	if formID == "" {
		return nil
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		return fmt.Errorf("connection ID parsing error: %w", err)
	}

	v := sdktypes.NewVar(vars.FormID, formID, false)
	v = v.WithScopeID(sdktypes.NewVarScopeID(cid))

	if err := h.vars.Set(ctx, v); err != nil {
		return err
	}
	return nil
}

func oauthURL(c sdkintegrations.ConnectionInit, form url.Values) string {
	urlPath := "/oauth/start/google%s?cid=%s&origin=%s"

	// Narrow down the requested scopes?
	oauthScopes := form.Get("auth_scopes")

	// Remember the AutoKitteh connection ID and connection origin.
	return fmt.Sprintf(urlPath, oauthScopes, c.ConnectionID, c.Origin)
}
