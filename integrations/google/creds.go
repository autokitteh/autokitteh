package google

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/google/calendar"
	"go.autokitteh.dev/autokitteh/integrations/google/drive"
	"go.autokitteh.dev/autokitteh/integrations/google/forms"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleCreds saves a new AutoKitteh connection with a user-submitted JSON key.
// It also acts as a passthrough for the OAuth connection mode, to save optional
// details (e.g. Google Form ID), to support and manage incoming events.
func (h handler) handleCreds(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check the "Content-Type" header.
	if common.PostWithoutFormContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("save connection: unexpected POST content type", zap.String("content_type", ct))
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	// Special case: save Google Calendar ID.
	calID := r.FormValue("cal_id")
	if calID != "" {
		if err := h.saveCalendarID(r.Context(), c, calID); err != nil {
			l.Error("Google Calendar ID saving error", zap.Error(err))
			c.AbortServerError("calendar ID saving error")
			return
		}
	}

	// Special case: validate & save Google Forms ID.
	formID := r.FormValue("form_id")
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

	switch r.FormValue("auth_type") {
	// GCP service-account JSON-key connection? Save the JSON key.
	case "", "json":
		ctx := extrazap.AttachLoggerToContext(l, r.Context())
		vs := sdktypes.EncodeVars(&vars.Vars{
			JSON:       r.PostFormValue("json"),
			CalendarID: calID,
			FormID:     formID,
		})
		h.finalize(ctx, c, vs.Set(vars.AuthType, "jsonKey", false))

	// User OAuth connect? Redirect to AutoKitteh's OAuth starting point.
	case "oauth":
		http.Redirect(w, r, oauthURL(r.Form, c), http.StatusFound)

	// Unknown mode.
	default:
		l.Error("Unexpected auth type", zap.String("auth_type", r.FormValue("auth_type")))
		c.AbortServerError(fmt.Sprintf("unexpected auth type %q", r.FormValue("auth_type")))
	}
}

func (h handler) saveCalendarID(ctx context.Context, c sdkintegrations.ConnectionInit, calID string) error {
	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		return fmt.Errorf("connection ID parsing error: %w", err)
	}

	v := sdktypes.NewVar(vars.CalendarID).SetValue(calID).WithScopeID(sdktypes.NewVarScopeID(cid))
	if err := h.vars.Set(ctx, v); err != nil {
		return err
	}
	return nil
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

// restoreJSONKey updates the JSON key value after a connection setup fails.
// If a previous JSON key existed, it's restored; otherwise, it's set to empty.
func (h handler) restoreJSONKey(ctx context.Context, cid sdktypes.ConnectionID, prevKey string) error {
	v := sdktypes.NewVar(vars.JSON).SetValue(prevKey).SetSecret(true).WithScopeID(sdktypes.NewVarScopeID(cid))
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

	prevVs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		l.Error("Previous JSON key retrieval error", zap.Error(err))
		c.AbortServerError("previous JSON key retrieval error")
		return
	}
	prevKey := prevVs.GetValue(vars.JSON)

	// Unique step for Google integrations (specifically Calendar, Forms, and
	// Gmail): save the auth data before creating/updating event watches.
	vsl := kittehs.TransformMapToList(vs.ToMap(), func(_ sdktypes.Symbol, v sdktypes.Var) sdktypes.Var {
		return v.WithScopeID(sdktypes.NewVarScopeID(cid))
	})

	if err := h.vars.Set(ctx, vsl...); err != nil {
		l.Error("Connection data saving error", zap.Error(err))
		c.AbortServerError("connection data saving error")
		return
	}

	if err := calendar.UpdateWatches(ctx, h.vars, cid); err != nil {
		h.handleWatchError(ctx, c, cid, prevKey, err, "Calendar")
		return
	}

	if err := drive.UpdateWatches(ctx, h.vars, cid); err != nil {
		h.handleWatchError(ctx, c, cid, prevKey, err, "Drive")
		return
	}

	if err := forms.UpdateWatches(ctx, h.vars, cid); err != nil {
		h.handleWatchError(ctx, c, cid, prevKey, err, "Forms")
		return
	}

	if err := gmail.UpdateWatch(ctx, h.vars, cid); err != nil {
		h.handleWatchError(ctx, c, cid, prevKey, err, "Gmail")
		return
	}

	// Redirect to the post-init handler to finish the connection setup.
	c.Finalize(vs)
}

func (h handler) handleWatchError(ctx context.Context, c sdkintegrations.ConnectionInit, cid sdktypes.ConnectionID, prevKey string, err error, watchName string) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	l.Error(fmt.Sprintf("Google %s watches creation error", watchName), zap.Error(err))
	c.AbortServerError(watchName + " watches creation error")
	jsonErr := h.restoreJSONKey(ctx, cid, prevKey)
	if jsonErr != nil {
		l.Error("JSON key deletion error", zap.Error(jsonErr))
		c.AbortServerError("JSON key deletion error")
	}
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
