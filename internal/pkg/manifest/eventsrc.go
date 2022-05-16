package manifest

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
)

type EventSource struct {
	ID       apieventsrc.EventSourceID `json:"id"`
	Disabled bool                      `json:"disabled"`
	Types    []string                  `json:"types"`
}

func (a EventSource) API() (*apieventsrc.EventSource, error) {
	return apieventsrc.NewEventSource(
		apieventsrc.EventSourceID(a.ID),
		(&apieventsrc.EventSourceSettings{}).
			SetTypes(a.Types).
			SetEnabled(!a.Disabled),
		time.Now(),
		nil,
	)
}

func (a EventSource) Compile() ([]*Action, error) {
	api, err := a.API()
	if err != nil {
		return nil, fmt.Errorf("invalid eventsrc: %w", err)
	}

	return []*Action{{
		Desc: fmt.Sprintf("create eventsrc %s", api.ID()),
		Run: func(ctx context.Context, env *Env) (string, error) {
			err := env.EventSources.Add(ctx, api.ID(), api.Settings())
			if err != nil {
				return "failed", err
			}

			return "created", nil
		},
	}}, nil
}
