package manifest

import (
	"context"
	"fmt"
	"time"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
)

type EventSource struct {
	ID       apieventsrc.EventSourceID `json:"id"`
	Disabled bool                      `json:"disabled"`
	Types    []string                  `json:"types"`
}

func (a EventSource) API(id string) (*apieventsrc.EventSource, error) {
	if a.ID != "" {
		id = a.ID.String()
	}

	return apieventsrc.NewEventSource(
		apieventsrc.EventSourceID(id),
		(&apieventsrc.EventSourceSettings{}).
			SetTypes(a.Types).
			SetEnabled(!a.Disabled),
		time.Now(),
		nil,
	)
}

func (a EventSource) Compile(id string) ([]*Action, error) {
	api, err := a.API(id)
	if err != nil {
		return nil, fmt.Errorf("invalid eventsrc: %w", err)
	}

	return []*Action{{
		Desc: fmt.Sprintf("create eventsrc %q", api.ID()),
		Run: func(ctx context.Context, env *Env) (string, error) {
			err := env.EventSources.Add(ctx, api.ID(), api.Settings())
			if err != nil {
				return "failed", err
			}

			return "created", nil
		},
	}}, nil
}
