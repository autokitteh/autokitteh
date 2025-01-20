package crontab

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

func (ct *Crontab) renewJiraEventWatchesWorkflow(wctx workflow.Context) error {
	// Enumerate all Jira connections (there's no single connection var value
	// that we're looking for, so we can't use "ct.vars.FindConnectionIDs").
	ctx := authcontext.SetAuthnSystemUser(context.Background())
	cs, err := ct.connections.List(ctx, sdkservices.ListConnectionsFilter{IntegrationID: jira.IntegrationID})
	if err != nil {
		ct.logger.Error("Failed to list Jira connections for event watch renewal", zap.Error(err))
		return err
	}

	// Check each Jira connection, renew its event watch if needed, abort on the first error.
	errs := make([]error, 0)
	for _, c := range cs {
		actx := temporalclient.WithActivityOptions(wctx, taskQueueName, ct.cfg.Activity)
		err := workflow.ExecuteActivity(actx, ct.renewJiraEventWatchActivity, c.ID()).Get(wctx, nil)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (ct *Crontab) renewJiraEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	l := ct.logger.With(zap.String("connection_id", cid.String()))
	ctx = authcontext.SetAuthnSystemUser(ctx)

	// Load the Jira connection's webhook expiration and ID.
	vs, err := ct.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		l.Error("Failed to get Jira connection vars for event watch renewal", zap.Error(err))
		return err
	}

	watchID := vs.GetValue(jira.WebhookID)
	if watchID == "" {
		// Jira connection uses an API token or a PAT instead of OAuth,
		// so it doesn't have an event watch to renew.
		return nil
	}

	id, err := strconv.Atoi(watchID)
	if err != nil {
		l.Error("Invalid Jira event watch ID for renewal", zap.Error(err))
		return err
	}

	twoWeeksFromNow := time.Now().UTC().AddDate(0, 0, 14)
	expiration, err := time.Parse(time.RFC3339, vs.GetValue(jira.WebhookExpiration))
	if err == nil && expiration.UTC().After(twoWeeksFromNow) {
		// The event watch is still valid for at least 2 weeks,
		// so there's no need to renew it.
		return nil
	}

	// Load the Jira OAuth configuration, to get a fresh OAuth access token.
	cfg, _, err := ct.oauth.Get(ctx, "jira")
	if err != nil {
		l.Error("Failed to get OAuth config for Jira event watch renewal", zap.Error(err))
		return err
	}

	refreshToken := vs.GetValueByString("oauth_RefreshToken")
	t, err := cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		l.Error("Failed to refresh OAuth token for Jira event watch renewal", zap.Error(err))
		return err
	}

	// Refresh the event watch (2 weeks --> 1 month).
	u, err := jira.APIBaseURL()
	if err != nil {
		l.Error("Invalid Atlassian base URL for Jira event watch renewal", zap.Error(err))
		return err
	}

	u, err = url.JoinPath(u, "/ex/jira", vs.GetValue(jira.AccessID))
	if err != nil {
		l.Error("Failed to construct Jira API URL for event watch renewal", zap.Error(err))
		return err
	}

	newExp, ok := jira.ExtendWebhookLife(l, u, t.AccessToken, id)
	if !ok {
		l.Error("Failed to renew Jira event watch")
		return fmt.Errorf("Failed to renew Jira event watch: %s", cid.String())
	}

	// Update the connection vars.
	vs.Set(sdktypes.NewSymbol("oauth_AccessToken"), t.AccessToken, true)
	vs.Set(sdktypes.NewSymbol("oauth_Expiry"), t.Expiry.String(), false)
	vs.Set(sdktypes.NewSymbol("oauth_RefreshToken"), t.RefreshToken, true)
	vs.Set(sdktypes.NewSymbol("WebhookExpiration"), newExp.Format(time.RFC3339), false)
	if err := ct.vars.Set(ctx, vs...); err != nil {
		l.Error("Failed to update Jira connection vars after event watch renewal", zap.Error(err))
		return err
	}
	return nil
}
