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
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	// Special case: validate & save Google Forms ID.
	formID := r.PostFormValue("form_id")
	if formID != "" {
		ok, err := regexp.MatchString(`[\w-]{20,}`, formID)
		if err != nil {
			l.Error("Google Forms form ID validation error", zap.Error(err))
			c.AbortServerError(fmt.Sprintf("form ID validation error: %v", err))
			return
		}
		if !ok {
			l.Warn("Invalid Google Forms form ID", zap.String("formID", formID))
			c.AbortBadRequest(fmt.Sprintf("invalid Google Forms ID %q", formID))
			return
		}

		if err := h.saveFormID(r.Context(), c, formID); err != nil {
			l.Error("Google Forms form ID saving error", zap.Error(err))
			c.AbortServerError("form ID saving error")
			return
		}
	}

	switch r.PostFormValue("auth_type") {
	// GCP service-account JSON-key connection? Save the JSON key.
	case "", "json":
		ctx := extrazap.AttachLoggerToContext(l, r.Context())
		vs := sdktypes.EncodeVars(&vars.Vars{JSON: r.PostFormValue("json"), FormID: formID})
		h.finalize(ctx, c, vs)

	// User OAuth connect? Redirect to AutoKitteh's OAuth starting point.
	case "oauth":
		http.Redirect(w, r, oauthURL(r.PostForm, c), http.StatusFound)

	// Unknown mode.
	default:
		l.Error("Unexpected auth type", zap.String("auth_type", r.PostFormValue("auth_type")))
		c.AbortServerError(fmt.Sprintf("unexpected auth type %q", r.PostFormValue("auth_type")))
	}
}

func (h handler) saveFormID(ctx context.Context, c sdkintegrations.ConnectionInit, formID string) error {
	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		return fmt.Errorf("connection ID parsing error: %w", err)
	}

	v := sdktypes.NewVar(vars.FormID).SetValue(formID).WithScopeID(sdktypes.NewVarScopeID(cid))
	if err := h.vars.Set(ctx, v); err != nil {
		return err
	}
	return nil
}

// finalize saves the user-submitted JSON key and optional Google Forms ID.
// It also initializes watches for Gmail and Google Forms, if needed.
func (h handler) finalize(ctx context.Context, c sdkintegrations.ConnectionInit, vs sdktypes.Vars) {
	l := extrazap.ExtractLoggerFromContext(ctx)

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("Invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	// Unique step for Google integrations (specifically for Gmail and Forms):
	// save the auth data before creating/updating event watches.
	vsl := kittehs.TransformMapToList(vs.ToMap(), func(_ sdktypes.Symbol, v sdktypes.Var) sdktypes.Var {
		return v.WithScopeID(sdktypes.NewVarScopeID(cid))
	})

	if err := h.vars.Set(ctx, vsl...); err != nil {
		l.Error("Connection data saving error", zap.Error(err))
		c.AbortServerError("connection data saving error")
		return
	}

	if err := forms.UpdateWatches(ctx, h.vars, cid); err != nil {
		l.Error("Google Forms watches creation error", zap.Error(err))
		c.AbortServerError("form watches creation error")
		return
	}

	if err := gmail.UpdateWatch(ctx, h.vars, cid); err != nil {
		l.Error("Gmail watch creation error", zap.Error(err))
		c.AbortServerError("Gmail watch creation error")
		return
	}

	// Redirect to the post-init handler to finish the connection setup.
	c.Finalize(vsl)
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
