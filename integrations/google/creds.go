package google

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/google/forms"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// handleCreds saves a new AutoKitteh connection with a user-submitted JSON key.
// It also acts as a passthrough for the OAuth connection mode, to save optional
// details (e.g. Google Form ID), to support and manage incoming events.
func (h handler) handleCreds(w http.ResponseWriter, r *http.Request) {
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

	// Special case: validate & save Google Forms ID.
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

		if err := h.saveFormID(r.Context(), c, formID); err != nil {
			l.Error("Form ID saving error", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError, "form ID saving error")
			return
		}
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("Invalid connection ID", zap.Error(err))
		c.Abort("invalid connection ID")
		return
	}

	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	switch r.PostFormValue("auth_type") {
	// GCP service-account JSON-key connection? Save the JSON key.
	case "", "json":
		c.Finalize(sdktypes.EncodeVars(&vars.Vars{JSON: r.PostFormValue("json"), FormID: formID}))

		if err := forms.UpdateWatches(ctx, h.vars, cid); err != nil {
			l.Error("Google form watches creation error", zap.Error(err))
		}

		if err := gmail.UpdateWatch(ctx, h.vars, cid); err != nil {
			l.Error("Gmail watch creation error", zap.Error(err))
		}

	// User OAuth connect? Redirect to AutoKitteh's OAuth starting point.
	case "oauth":
		http.Redirect(w, r, oauthURL(r.PostForm, c), http.StatusFound)

	// Unknown mode.
	default:
		err := fmt.Sprintf("unexpected auth type %q", r.PostFormValue("auth_type"))
		c.AbortWithStatus(http.StatusInternalServerError, err)
	}
}

func (h handler) saveFormID(ctx context.Context, c sdkintegrations.ConnectionInit, formID string) error {
	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		return fmt.Errorf("connection ID parsing error: %w", err)
	}

	v := sdktypes.NewVar(vars.FormID, formID, false).WithScopeID(sdktypes.NewVarScopeID(cid))
	if err := h.vars.Set(ctx, v); err != nil {
		return err
	}
	return nil
}

func oauthURL(form url.Values, c sdkintegrations.ConnectionInit) string {
	// Default scopes for OAuth: all ("google").
	u := "/oauth/start/google%s?cid=%s&origin=%s"

	// Narrow down the requested scopes?
	oauthScopes := form.Get("auth_scopes")

	// Remember the AutoKitteh connection ID and connection origin.
	u = fmt.Sprintf(u, oauthScopes, c.ConnectionID, c.Origin)

	// Special case: "gmail" is the only one without a "google" prefix.
	return strings.Replace(u, "/googlegmail?", "/gmail?", 1)
}
