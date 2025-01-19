package calendar

import (
	"context"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/api/calendar/v3"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// UpdateWatches creates or renews calendar watches for a specific
// Google Calendar, if an ID was specified during initialization.
func UpdateWatches(ctx context.Context, v sdkservices.Vars, connID sdktypes.ConnectionID) error {
	l := extrazap.ExtractLoggerFromContext(ctx)

	a := api{logger: l, vars: v, cid: connID}
	calID, err := a.calendarID(ctx)
	if err != nil {
		return err
	}

	// No calendar ID? Nothing to do.
	if calID == "" {
		l.Debug("No calendar ID specified, skipping Google Calendar watches")
		return nil
	}

	// Get the user's email address.
	vs, err := v.Get(ctx, sdktypes.NewVarScopeID(connID), vars.UserEmail)
	if err != nil {
		return err
	}
	userEmail := vs.Get(vars.UserEmail).Value()

	// Create or renew the events watch channel.
	l = l.With(zap.String("calendarID", calID))
	extrazap.AttachLoggerToContext(l, ctx)
	watchChannel, err := a.watchEvents(ctx, connID, userEmail, calID)
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
	err = a.syncEvents(ctx, calID)
	if err != nil {
		return err
	}

	return nil
}

func (a api) saveWatchChannel(ctx context.Context, cid sdktypes.ConnectionID, wc *calendar.Channel) error {
	expiration := strconv.FormatInt(wc.Expiration, 10)
	sid := sdktypes.NewVarScopeID(cid)

	vs := sdktypes.NewVars().
		Set(vars.CalendarEventsWatchID, wc.Id, false).
		Set(vars.CalendarEventsWatchResID, wc.ResourceId, false).
		Set(vars.CalendarEventsWatchExp, expiration, false)

	for _, v := range vs {
		if err := a.vars.Set(ctx, v.WithScopeID(sid)); err != nil {
			return err
		}
	}

	return nil
}
