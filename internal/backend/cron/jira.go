package cron

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

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
		return true // Old/invalid expiration time --> solve by renewing the watch.
	}

	// Update this event watch only if it's about to expire in less than 2 weeks.
	twoWeeksFromNow := time.Now().UTC().AddDate(0, 0, 14)
	return t.UTC().Before(twoWeeksFromNow)
}

func (cr *Cron) renewJiraEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	l := cr.logger.With(zap.String("connection_id", cid.String()))
	ctx = authcontext.SetAuthnSystemUser(ctx)

	// Load the Jira connection's webhook ID.
	vs, err := cr.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		l.Error("failed to get Jira connection vars for event watch renewal", zap.Error(err))
		return err
	}

	id, err := strconv.Atoi(vs.GetValue(jira.WebhookID))
	if err != nil {
		l.Error("invalid Jira event watch ID for renewal", zap.Error(err))
		return err
	}

	// Load the Jira OAuth configuration, to get a fresh OAuth access token.
	cfg, _, err := cr.oauth.Get(ctx, "jira")
	if err != nil {
		l.Error("failed to get OAuth config for Jira event watch renewal", zap.Error(err))
		return err
	}

	refreshToken := vs.GetValueByString("oauth_RefreshToken")
	t, err := cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		l.Error("failed to refresh OAuth token for Jira event watch renewal", zap.Error(err))
		return err
	}

	// Refresh the event watch (2 weeks --> 1 month).
	u, err := jira.APIBaseURL()
	if err != nil {
		l.Error("invalid Atlassian base URL for Jira event watch renewal", zap.Error(err))
		return err
	}

	u, err = url.JoinPath(u, "/ex/jira", vs.GetValue(jira.AccessID))
	if err != nil {
		l.Error("failed to construct Jira API URL for event watch renewal", zap.Error(err))
		return err
	}

	newExp, ok := jira.ExtendWebhookLife(l, u, t.AccessToken, id)
	if !ok {
		l.Error("failed to renew Jira event watch")
		return fmt.Errorf("failed to renew Jira event watch: %s", cid.String())
	}

	// Update the connection vars.
	vs.Set(sdktypes.NewSymbol("oauth_AccessToken"), t.AccessToken, true)
	vs.Set(sdktypes.NewSymbol("oauth_Expiry"), t.Expiry.String(), false)
	vs.Set(sdktypes.NewSymbol("oauth_RefreshToken"), t.RefreshToken, true)
	vs.Set(sdktypes.NewSymbol("WebhookExpiration"), newExp.Format(time.RFC3339), false)
	if err := cr.vars.Set(ctx, vs...); err != nil {
		l.Error("failed to update Jira connection vars after event watch renewal", zap.Error(err))
		return err
	}

	return nil
}
