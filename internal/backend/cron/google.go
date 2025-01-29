package cron

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/google/calendar"
	"go.autokitteh.dev/autokitteh/integrations/google/drive"
	"go.autokitteh.dev/autokitteh/integrations/google/forms"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/vars"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO(INT-190): Move this to the Google integration package.

type (
	listActivity  func(context.Context) ([]sdktypes.ConnectionID, error)
	renewActivity func(context.Context, sdktypes.ConnectionID) error
)

func (cr *Cron) renewGoogleEventWatchesWorkflow(wctx workflow.Context, la listActivity, ra renewActivity) error {
	actx := temporalclient.WithActivityOptions(wctx, taskQueueName, cr.cfg.Activity)

	var cids []sdktypes.ConnectionID
	if err := workflow.ExecuteActivity(actx, la).Get(wctx, &cids); err != nil {
		return err
	}

	var errs []error
	for _, cid := range cids {
		if err := workflow.ExecuteActivity(actx, ra, cid).Get(wctx, nil); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (cr *Cron) renewGmailEventWatchesWorkflow(wctx workflow.Context) error {
	return cr.renewGoogleEventWatchesWorkflow(wctx,
		cr.listGmailConnectionsActivity,
		cr.renewGmailEventWatchActivity,
	)
}

func (cr *Cron) renewGoogleCalendarEventWatchesWorkflow(wctx workflow.Context) error {
	return cr.renewGoogleEventWatchesWorkflow(wctx,
		cr.listGoogleCalendarConnectionsActivity,
		cr.renewGoogleCalendarEventWatchActivity,
	)
}

func (cr *Cron) renewGoogleDriveEventWatchesWorkflow(wctx workflow.Context) error {
	return cr.renewGoogleEventWatchesWorkflow(wctx,
		cr.listGoogleDriveConnectionsActivity,
		cr.renewGoogleDriveEventWatchActivity,
	)
}

func (cr *Cron) renewGoogleFormsEventWatchesWorkflow(wctx workflow.Context) error {
	return cr.renewGoogleEventWatchesWorkflow(wctx,
		cr.listGoogleFormsConnectionsActivity,
		cr.renewGoogleFormsEventWatchesActivity,
	)
}

func (cr *Cron) listGmailConnectionsActivity(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	ctx = authcontext.SetAuthnSystemUser(ctx)

	// Enumerate all Gmail connections (there's no single connection var value
	// that we're looking for, so we can't use "cr.vars.FindConnectionIDs").
	cs, err := cr.connections.List(ctx, sdkservices.ListConnectionsFilter{
		IntegrationID: gmail.IntegrationID,
	})
	if err != nil {
		cr.logger.Error("failed to list Gmail connections for event watch renewal", zap.Error(err))
		return nil, err
	}

	var cids []sdktypes.ConnectionID
	for _, c := range cs {
		cid := c.ID()
		if cr.checkGmailEventWatch(ctx, cid) {
			cids = append(cids, cid)
		}
	}

	return cids, nil
}

func (cr *Cron) listGoogleCalendarConnectionsActivity(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	ctx = authcontext.SetAuthnSystemUser(ctx)

	// Enumerate all Google Calendar connections (there's no single connection var
	// value that we're looking for, so we can't use "cr.vars.FindConnectionIDs").
	cs, err := cr.connections.List(ctx, sdkservices.ListConnectionsFilter{
		IntegrationID: calendar.IntegrationID,
	})
	if err != nil {
		cr.logger.Error("failed to list Google Calendar connections for event watch renewal", zap.Error(err))
		return nil, err
	}

	var cids []sdktypes.ConnectionID
	for _, c := range cs {
		cid := c.ID()
		if cr.checkGoogleCalendarEventWatch(ctx, cid) {
			cids = append(cids, cid)
		}
	}

	return cids, nil
}

func (cr *Cron) listGoogleDriveConnectionsActivity(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	ctx = authcontext.SetAuthnSystemUser(ctx)

	// Enumerate all Google Drive connections (there's no single connection var value
	// that we're looking for, so we can't use "cr.vars.FindConnectionIDs").
	cs, err := cr.connections.List(ctx, sdkservices.ListConnectionsFilter{
		IntegrationID: drive.IntegrationID,
	})
	if err != nil {
		cr.logger.Error("failed to list Google Drive connections for event watch renewal", zap.Error(err))
		return nil, err
	}

	var cids []sdktypes.ConnectionID
	for _, c := range cs {
		cid := c.ID()
		if cr.checkGoogleDriveEventWatch(ctx, cid) {
			cids = append(cids, cid)
		}
	}

	return cids, nil
}

func (cr *Cron) listGoogleFormsConnectionsActivity(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	ctx = authcontext.SetAuthnSystemUser(ctx)

	// Enumerate all Google Forms connections (there's no single connection var
	// value that we're looking for, so we can't use "cr.vars.FindConnectionIDs").
	cs, err := cr.connections.List(ctx, sdkservices.ListConnectionsFilter{
		IntegrationID: forms.IntegrationID,
	})
	if err != nil {
		cr.logger.Error("failed to list Google Forms connections for event watch renewal", zap.Error(err))
		return nil, err
	}

	var cids []sdktypes.ConnectionID
	for _, c := range cs {
		cid := c.ID()
		if cr.checkGoogleFormsEventWatch(ctx, cid) {
			cids = append(cids, cid)
		}
	}

	return cids, nil
}

func (cr *Cron) checkGmailEventWatch(ctx context.Context, cid sdktypes.ConnectionID) bool {
	l := cr.logger.With(zap.String("connection_id", cid.String()))

	vs, err := cr.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.GmailWatchExpiration)
	if err != nil {
		l.Error("failed to get Gmail connection vars for event watch renewal", zap.Error(err))
		return false
	}

	e := vs.GetValue(vars.GmailWatchExpiration)
	t, err := time.Parse(time.RFC3339, e)
	if err != nil {
		l.Warn("invalid Gmail event watch expiration time during renewal check",
			zap.String("expiration", e), zap.Error(err),
		)
		return false
	}

	// Update this event watch only if it's about to expire in less than 3 days.
	threeDaysFromNow := time.Now().UTC().AddDate(0, 0, 3)
	return t.UTC().Before(threeDaysFromNow)
}

func (cr *Cron) checkGoogleCalendarEventWatch(ctx context.Context, cid sdktypes.ConnectionID) bool {
	l := cr.logger.With(zap.String("connection_id", cid.String()))

	vs, err := cr.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.CalendarEventsWatchExp)
	if err != nil {
		l.Error("failed to get Google Drive connection vars for event watch renewal", zap.Error(err))
		return false
	}

	e := vs.GetValue(vars.CalendarEventsWatchExp)
	timestamp, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		l.Warn("invalid Google Calendar event watch expiration timestamp",
			zap.String("expiration", e), zap.Error(err),
		)
		return false
	}

	// Update this event watch only if it's about to expire in less than 3 days.
	threeDaysFromNow := time.Now().UTC().AddDate(0, 0, 3)
	return time.Unix(timestamp/1000, 0).UTC().Before(threeDaysFromNow)
}

func (cr *Cron) checkGoogleDriveEventWatch(ctx context.Context, cid sdktypes.ConnectionID) bool {
	l := cr.logger.With(zap.String("connection_id", cid.String()))

	vs, err := cr.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.DriveEventsWatchExp)
	if err != nil {
		l.Error("failed to get Google Drive connection vars for event watch renewal", zap.Error(err))
		return false
	}

	e := vs.GetValue(vars.DriveEventsWatchExp)
	timestamp, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		l.Warn("invalid Google Drive event watch expiration timestamp",
			zap.String("expiration", e), zap.Error(err),
		)
		return false
	}

	// Update this event watch only if it's about to expire in less than 3 days.
	threeDaysFromNow := time.Now().UTC().AddDate(0, 0, 3)
	return time.Unix(timestamp, 0).UTC().Before(threeDaysFromNow)
}

func (cr *Cron) checkGoogleFormsEventWatch(ctx context.Context, cid sdktypes.ConnectionID) bool {
	l := cr.logger.With(zap.String("connection_id", cid.String()))

	vs, err := cr.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.FormWatchesExpiration)
	if err != nil {
		l.Error("failed to get Google Forms connection vars for event watch renewal", zap.Error(err))
		return false
	}

	e := vs.GetValue(vars.FormWatchesExpiration)
	t, err := time.Parse(time.RFC3339, e)
	if err != nil {
		l.Warn("invalid Google Forms event watch expiration time during renewal check",
			zap.String("expiration", e), zap.Error(err),
		)
		return false
	}

	// Update this event watch only if it's about to expire in less than 3 days.
	threeDaysFromNow := time.Now().UTC().AddDate(0, 0, 3)
	return t.UTC().Before(threeDaysFromNow)
}

type update func(context.Context, sdkservices.Vars, sdktypes.ConnectionID) error

func (cr *Cron) renewGoogleEventWatchesActivity(ctx context.Context, cid sdktypes.ConnectionID, u update) error {
	l := cr.logger.With(zap.String("connection_id", cid.String()))
	ctx = authcontext.SetAuthnSystemUser(ctx)

	err := u(ctx, cr.vars, cid)
	if err != nil {
		l.Error("failed to renew Google event watches", zap.Error(err))
		return err
	}

	return nil
}

func (cr *Cron) renewGmailEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	return cr.renewGoogleEventWatchesActivity(ctx, cid, gmail.UpdateWatch)
}

func (cr *Cron) renewGoogleCalendarEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	return cr.renewGoogleEventWatchesActivity(ctx, cid, calendar.UpdateWatches)
}

func (cr *Cron) renewGoogleDriveEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	return cr.renewGoogleEventWatchesActivity(ctx, cid, drive.UpdateWatches)
}

func (cr *Cron) renewGoogleFormsEventWatchesActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	return cr.renewGoogleEventWatchesActivity(ctx, cid, forms.UpdateWatches)
}
