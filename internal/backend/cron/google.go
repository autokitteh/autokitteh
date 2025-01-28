package cron

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/google/drive"
	"go.autokitteh.dev/autokitteh/integrations/google/vars"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO(INT-190): Move this to the Google integration package.

func (cr *Cron) renewGoogleCalendarEventWatchesWorkflow(wctx workflow.Context) error {
	return nil // TODO(INT-184): Implement.
}

func (cr *Cron) renewGoogleDriveEventWatchesWorkflow(wctx workflow.Context) error {
	actx := temporalclient.WithActivityOptions(wctx, taskQueueName, cr.cfg.Activity)

	var cids []sdktypes.ConnectionID
	err := workflow.ExecuteActivity(actx, cr.listGoogleDriveConnectionsActivity).Get(wctx, &cids)
	if err != nil {
		return err
	}

	errs := make([]error, 0)
	for _, cid := range cids {
		err := workflow.ExecuteActivity(actx, cr.renewGoogleDriveEventWatchActivity, cid).Get(wctx, nil)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (cr *Cron) listGoogleCalendarConnectionsActivity(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	return nil, nil // TODO(INT-184): Implement.
}

func (cr *Cron) listGoogleDriveConnectionsActivity(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	ctx = authcontext.SetAuthnSystemUser(ctx)

	// Enumerate all Google Drive connections (there's no single connection var value
	// that we're looking for, so we can't use "ct.vars.FindConnectionIDs").
	cs, err := cr.connections.List(ctx, sdkservices.ListConnectionsFilter{
		IntegrationID: drive.IntegrationID,
	})
	if err != nil {
		cr.logger.Error("failed to list Google Drive connections for event watch renewal", zap.Error(err))
		return nil, err
	}

	cids := make([]sdktypes.ConnectionID, 0)
	for _, c := range cs {
		cid := c.ID()
		if cr.checkGoogleDriveEventWatch(ctx, cid) {
			cids = append(cids, cid)
		}
	}

	return cids, nil
}

func (cr *Cron) checkGoogleDriveEventWatch(ctx context.Context, cid sdktypes.ConnectionID) bool {
	l := cr.logger.With(zap.String("connection_id", cid.String()))

	vs, err := cr.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.DriveEventsWatchExp)
	if err != nil {
		l.Error("failed to get Google Drive connection vars for event watch renewal", zap.Error(err))
		return false
	}

	e := vs.GetValue(vars.DriveEventsWatchExp)
	threeDaysFromNow := time.Now().UTC().AddDate(0, 0, 3)

	timestamp, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		l.Error("invalid Google Drive event watch expiration timestamp",
			zap.String("expiration", e), zap.Error(err),
		)
		return false
	}

	return time.Unix(timestamp, 0).UTC().Before(threeDaysFromNow)
}

func (cr *Cron) renewGoogleCalendarEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	return nil // TODO(INT-184): Implement.
}

func (cr *Cron) renewGoogleDriveEventWatchActivity(ctx context.Context, cid sdktypes.ConnectionID) error {
	l := cr.logger.With(zap.String("connection_id", cid.String()))
	ctx = authcontext.SetAuthnSystemUser(ctx)

	err := drive.UpdateWatches(ctx, cr.vars, cid)
	if err != nil {
		l.Error("failed to renew Google Drive event watch", zap.Error(err))
		return err
	}

	return nil
}
