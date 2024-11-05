package drive

import (
	"context"
	"strconv"
	"strings"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
	"google.golang.org/api/drive/v3"
)

// UpdateWatches creates or renews watches for a specific Google Drive
// or shared drive, if an ID was specified during initialization.
func UpdateWatches(ctx context.Context, v sdkservices.Vars, connID sdktypes.ConnectionID) error {
	l := extrazap.ExtractLoggerFromContext(ctx)

	// Not a Google Drive user? Nothing to do.
	if !gotDriveScope(ctx, v, connID) {
		l.Debug("No Google Drive scope, skipping Google Drive watches")
		return nil
	}

	a := api{logger: l, vars: v, cid: connID}
	// TODO: fileID? Should we allow users to specify both fileID and driveID?
	driveID, err := a.driveID(ctx)
	if err != nil {
		return err
	}

	// Get the user's email address.
	vs, err := v.Get(ctx, sdktypes.NewVarScopeID(connID), vars.UserEmail)
	if err != nil {
		return err
	}
	userEmail := vs.Get(vars.UserEmail).Value()

	// Create or renew the events watch channel.
	l = l.With(zap.String("driveID", driveID))
	extrazap.AttachLoggerToContext(l, ctx)
	watchChannel, err := a.watchEvents(ctx, connID, userEmail, driveID)
	if err != nil {
		return err
	}

	// Save all of its IDs.
	err = a.saveWatchChannel(ctx, connID, watchChannel)
	if err != nil {
		return err
	}

	// Initialize the watch channel's sync token, for event enrichment
	// (https://developers.google.com/calendar/api/guides/sync).
	err = a.syncEvents(ctx, driveID)
	if err != nil {
		return err
	}

	return nil
}

func gotDriveScope(ctx context.Context, v sdkservices.Vars, cid sdktypes.ConnectionID) bool {
	l := extrazap.ExtractLoggerFromContext(ctx)

	vs, err := v.Get(ctx, sdktypes.NewVarScopeID(cid), vars.UserScope)
	if err != nil {
		l.Error("Failed to get Google OAuth scopes", zap.Error(err))
		return false
	}

	return strings.Contains(vs.GetValue(vars.UserScope), "https://www.googleapis.com/auth/drive")
}

func (a api) saveWatchChannel(ctx context.Context, cid sdktypes.ConnectionID, wc *drive.Channel) error {
	expiration := strconv.FormatInt(wc.Expiration, 10)
	sid := sdktypes.NewVarScopeID(cid)

	vs := sdktypes.NewVars().
		Set(vars.DriveEventsWatchID, wc.Id, false).
		Set(vars.DriveEventsWatchResID, wc.ResourceId, false).
		Set(vars.DriveEventsWatchExp, expiration, false)

	for _, v := range vs {
		if err := a.vars.Set(ctx, v.WithScopeID(sid)); err != nil {
			return err
		}
	}

	return nil
}
