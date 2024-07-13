package google

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/google/forms"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
)

// Create/renew event watches for a specific Google Forms form, if its ID was specified.
func (h handler) updateFormWatches(ctx context.Context, c sdkintegrations.ConnectionInit) error {
	l := extrazap.ExtractLoggerFromContext(ctx)
	api := forms.API{Vars: h.vars, CID: c.ConnectionID}

	formID, err := api.FormID(ctx)
	if err != nil {
		return err
	}

	l = l.With(zap.String("formID", formID))

	// No form ID? Nothing to do.
	if formID == "" {
		return nil
	}

	// List all existing watches.
	watches, err := api.WatchesList(ctx)
	if err != nil {
		return fmt.Errorf("failed to list form watches: %w", err)
	}
	ws := kittehs.ListToMap(watches, func(w *forms.Watch) (forms.WatchEventType, *forms.Watch) {
		return forms.WatchEventType(w.EventType), w
	})

	// Renew or create the form's SCHEMA (changes) and (new) RESPONSES watches.
	for _, e := range []forms.WatchEventType{forms.WatchSchemaChanges, forms.WatchNewResponses} {
		w, ok := ws[e]
		if ok {
			l.Info("Found form watch", zap.Any("watch", w))
			w, err = api.WatchesRenew(ctx, w.Id)
			if err != nil {
				return fmt.Errorf("failed to renew %s watch for form %q: %w", e, formID, err)
			}
			l.Info("Renewed form watch", zap.Any("watch", w))
		} else {
			w, err = api.WatchesCreate(ctx, e)
			if err != nil {
				return fmt.Errorf("failed to create %s watch for form %q: %w", e, formID, err)
			}
			l.Info("Created form watch", zap.Any("watch", w))
		}
	}

	return nil
}
