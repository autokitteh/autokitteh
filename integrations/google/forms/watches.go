package forms

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/api/forms/v1"

	"go.autokitteh.dev/autokitteh/integrations/google/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type WatchEventType string

const (
	WatchSchemaChanges WatchEventType = "SCHEMA"
	WatchNewResponses  WatchEventType = "RESPONSES"
)

// UpdateWatches creates or renews schema-changes and new-responses event watches
// for a specific Google Forms form, if an ID was specified during initialization.
func UpdateWatches(ctx context.Context, v sdkservices.Vars, cid sdktypes.ConnectionID) error {
	l := extrazap.ExtractLoggerFromContext(ctx)

	a := api{vars: v, cid: cid}
	formID, err := a.formID(ctx)
	if err != nil {
		return err
	}

	// No form ID? Nothing to do.
	if formID == "" {
		l.Debug("No form ID specified, skipping Google Forms watches")
		return nil
	}

	// List all existing watches.
	l = l.With(zap.String("formID", formID))
	extrazap.AttachLoggerToContext(l, ctx)
	watches, err := a.watchesList(ctx)
	if err != nil {
		return fmt.Errorf("failed to list watches for form ID %q: %w", formID, err)
	}
	ws := kittehs.ListToMap(watches, func(w *forms.Watch) (WatchEventType, *forms.Watch) {
		return WatchEventType(w.EventType), w
	})

	// Renew or create the form's SCHEMA (changes) and (new) RESPONSES watches.
	for _, e := range []WatchEventType{WatchSchemaChanges, WatchNewResponses} {
		watchID, err := a.updateSingleWatch(ctx, e, ws[e])
		if err != nil {
			return err
		}
		// And save their IDs.
		err = a.saveWatchID(ctx, e, watchID)
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

func (a api) saveWatchID(ctx context.Context, e WatchEventType, watchID string) error {
	n := vars.FormResponsesWatchID
	if e == WatchSchemaChanges {
		n = vars.FormSchemaWatchID
	}

	v := sdktypes.NewVar(n).SetValue(watchID).WithScopeID(sdktypes.NewVarScopeID(a.cid))
	if err := a.vars.Set(ctx, v); err != nil {
		return err
	}

	return nil
}
