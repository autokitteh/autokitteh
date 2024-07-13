package forms

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/api/forms/v1"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type WatchEventType string

const (
	WatchSchemaChanges WatchEventType = "SCHEMA"
	WatchNewResponses  WatchEventType = "RESPONSES"
)

// UpdateWatches creates or renews schema-changes and new-responses event watches
// for a specific Google Forms form, if an ID was specified during initialization.
func UpdateWatches(ctx context.Context, v sdkservices.Vars, c sdkintegrations.ConnectionInit) error {
	l := extrazap.ExtractLoggerFromContext(ctx)
	api := api{Vars: v, CID: c.ConnectionID}

	formID, err := api.formID(ctx)
	if err != nil {
		return err
	}

	l = l.With(zap.String("formID", formID))
	extrazap.AttachLoggerToContext(l, ctx)

	// No form ID? Nothing to do.
	if formID == "" {
		return nil
	}

	// List all existing watches.
	watches, err := api.watchesList(ctx)
	if err != nil {
		return fmt.Errorf("failed to list form watches: %w", err)
	}
	ws := kittehs.ListToMap(watches, func(w *forms.Watch) (WatchEventType, *forms.Watch) {
		return WatchEventType(w.EventType), w
	})

	// Renew or create the form's SCHEMA (changes) and (new) RESPONSES watches.
	for _, e := range []WatchEventType{WatchSchemaChanges, WatchNewResponses} {
		wid, err := api.updateSingleWatch(ctx, e, ws[e])
		if err != nil {
			return err
		}
		// And save their IDs.
		err = saveWatchID(ctx, v, e, wid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a api) updateSingleWatch(ctx context.Context, e WatchEventType, w *forms.Watch) (string, error) {
	l := extrazap.ExtractLoggerFromContext(ctx)

	// Create a new watch...
	if w == nil {
		w, err := a.watchesCreate(ctx, e)
		if err != nil {
			return "", fmt.Errorf("failed to create %s form watch: %w", e, err)
		}
		l.Info("Created form watch", zap.Any("watch", w))
		return w.Id, nil
	}

	// ...Or renew an existing watch.
	l.Info("Found existing form watch", zap.Any("watch", w))
	w, err := a.watchesRenew(ctx, w.Id)
	if err != nil {
		return "", fmt.Errorf("failed to renew %s form watch: %w", e, err)
	}
	l.Info("Renewed form watch", zap.Any("watch", w))
	return w.Id, nil
}

func saveWatchID(ctx context.Context, v sdkservices.Vars, e WatchEventType, wid string) error {
	// TODO(ENG-1103): Save the watch ID in the connection's vars.
	return nil
}
