package cron

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/atlassian/jira"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO(INT-190): Move this to the Jira integration package.

func (cr *Cron) renewJiraEventWatchesWorkflow(wctx workflow.Context) error {
	actx := temporalclient.WithActivityOptions(wctx, taskQueueName, cr.cfg.Activity)

	var cids []sdktypes.ConnectionID
	err := workflow.ExecuteActivity(actx, cr.listJiraConnectionsActivity).Get(wctx, &cids)
	if err != nil {
		return err
	}

	var errs []error
	for _, cid := range cids {
		err := workflow.ExecuteActivity(actx, cr.renewJiraEventWatchActivity, cid).Get(wctx, nil)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (cr *Cron) listJiraConnectionsActivity(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	ctx = authcontext.SetAuthnSystemUser(ctx)

	// Enumerate all Jira connections (there's no single connection var value
	// that we're looking for, so we can't use "cr.vars.FindConnectionIDs").
	cs, err := cr.connections.List(ctx, sdkservices.ListConnectionsFilter{
		IntegrationID: jira.IntegrationID,
	})
	if err != nil {
		cr.logger.Error("failed to list Jira connections for event watch renewal", zap.Error(err))
		return nil, err
	}

	var cids []sdktypes.ConnectionID
	for _, c := range cs {
		cid := c.ID()
		if cr.checkJiraEventWatch(ctx, cid) {
			cids = append(cids, cid)
		}
	}

	return cids, nil
}

func (cr *Cron) checkJiraEventWatch(ctx context.Context, cid sdktypes.ConnectionID) bool {
	l := cr.logger.With(zap.String("connection_id", cid.String()))

	vs, err := cr.vars.Get(ctx, sdktypes.NewVarScopeID(cid), jira.WebhookID, jira.WebhookExpiration)
	if err != nil {
		l.Error("failed to get Jira connection vars for event watch renewal", zap.Error(err))
		return false
	}

	watchID := vs.GetValue(jira.WebhookID)
	if watchID == "" {
		// Jira connection uses an API token or a PAT instead of OAuth,
		// so it doesn't have an event watch to renew.
		return false
	}

	e := vs.GetValue(jira.WebhookExpiration)
	t, err := time.Parse(time.RFC3339, e)
	if err != nil {
		l.Warn("invalid Jira event watch expiration time during renewal check",
			zap.String("expiration", e), zap.Error(err),
		)
		return true // Invalid expiration time: don't ignore watch, solve by renewing it.
	}

	// Update this event watch only if it's about to expire in less than 2 weeks.
	twoWeeksFromNow := time.Now().UTC().AddDate(0, 0, 14)
	return t.UTC().Before(twoWeeksFromNow)
}

func (cr *Cron) renewJiraEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	l := cr.logger.With(zap.String("connection_id", cid.String()))
	l.Debug("renewing Jira event watch in: " + cid.String())

	ctx = authcontext.SetAuthnSystemUser(ctx)
	vsid := sdktypes.NewVarScopeID(cid)

	// Load the Jira connection's webhook ID.
	vs, err := cr.vars.Get(ctx, vsid)
	if err != nil {
		l.Error("failed to get Jira connection vars for event watch renewal", zap.Error(err))
		return err
	}

	accessID := vs.GetValue(jira.AccessID)

	wid := vs.GetValue(jira.WebhookID)
	id, err := strconv.Atoi(wid)
	if err != nil {
		l.Warn("invalid Jira event watch ID for renewal", zap.String("watch_id", wid), zap.Error(err))
		cr.deleteInvalidWatchID(ctx, cid, wid)
		return nil // No need to retry.
	}

	// Update the Jira OAuth configuration, to get a fresh OAuth access token.
	cfg, _, err := cr.oauth.Get(ctx, "jira")
	if err != nil {
		l.Error("failed to get OAuth config for Jira event watch renewal", zap.Error(err))
		return err
	}

	t := &oauth2.Token{
		AccessToken:  vs.GetValueByString("oauth_AccessToken"),
		RefreshToken: vs.GetValueByString("oauth_RefreshToken"),
		TokenType:    vs.GetValueByString("oauth_TokenType"),
		Expiry:       parseTimeSafely(vs.GetValueByString("oauth_Expiry")),
	}

	if t.Expiry.UTC().Before(time.Now().UTC()) {
		t, err = cfg.TokenSource(ctx, t).Token()
		if err != nil {
			l.Error("failed to refresh OAuth token for Jira event watch renewal", zap.Error(err))
			return err
		}

		vs = sdktypes.NewVars(
			vs.Get(sdktypes.NewSymbol("oauth_AccessToken")).SetValue(t.AccessToken),
			vs.Get(sdktypes.NewSymbol("oauth_RefreshToken")).SetValue(t.RefreshToken),
			vs.Get(sdktypes.NewSymbol("oauth_Expiry")).SetValue(t.Expiry.String()),
		)
		if err = cr.vars.Set(ctx, vs...); err != nil {
			l.Error("failed to update Jira connection vars after OAuth token refresh", zap.Error(err))
			// We have a valid OAuth token, but we can't save it. This may cause problems
			// down the line, but for now we can at least try to renew the event watch.
		}
	}

	// Refresh the event watch (2 weeks --> 1 month).
	u, err := jira.APIBaseURL()
	if err != nil {
		l.Error("invalid Atlassian base URL for Jira event watch renewal", zap.Error(err))
		return temporal.NewNonRetryableApplicationError(
			"invalid Atlassian base URL: "+err.Error(), "URLParseError", err, cid.String(), wid,
		)
	}

	u, err = url.JoinPath(u, "/ex/jira", accessID)
	if err != nil {
		l.Error("failed to construct Jira API URL for event watch renewal", zap.Error(err))
		return temporal.NewNonRetryableApplicationError(
			"invalid Atlassian base URL: "+err.Error(), "URLParseError", err, cid.String(), wid,
		)
	}

	newExp, ok := jira.ExtendWebhookLife(l, u, t.AccessToken, id)
	if !ok {
		l.Error("failed to renew Jira event watch")
		return fmt.Errorf("failed to renew Jira event watch: %s", cid.String())
	}

	v := sdktypes.NewVar(jira.WebhookExpiration).SetValue(newExp.Format(time.RFC3339))
	if err := cr.vars.Set(ctx, v.WithScopeID(vsid)); err != nil {
		l.Error("failed to update Jira connection var after event watch renewal", zap.Error(err))
		return err
	}

	l.Debug("finished renewing Jira event watch in: " + cid.String())
	return nil
}

func (cr *Cron) deleteInvalidWatchID(ctx context.Context, cid sdktypes.ConnectionID, watchID string) {
	err := cr.vars.Delete(ctx, sdktypes.NewVarScopeID(cid), jira.WebhookID)
	if err != nil {
		cr.logger.Error("failed to delete invalid Jira event watch ID during renewal",
			zap.String("connection_id", cid.String()),
			zap.String("watch_id", watchID),
			zap.Error(err),
		)
	}
}

func parseTimeSafely(s string) time.Time {
	// Remove unnecessary suffixes, e.g. "PST m=+3759.281638293".
	s = regexp.MustCompile(`\s+[A-Z].*`).ReplaceAllString(s, "")

	// Go time format.
	if t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700", s); err == nil {
		return t
	}

	// RFC-3339.
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}

	// Fallback if we don't know the format: refresh the OAuth token.
	return time.Now().Add(-time.Hour)
}
