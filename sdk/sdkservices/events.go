package sdkservices

import (
	"context"
	"time"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListEventsFilter struct {
	IntegrationID     sdktypes.IntegrationID
	DestinationID     sdktypes.EventDestinationID
	EventType         string
	Limit             int
	CreatedAfter      *time.Time
	MinSequenceNumber uint64
	Order             ListOrder
}

type ListOrder string

const (
	ListOrderAscending  ListOrder = "ASC"
	ListOrderDescending ListOrder = "DESC"
)

type Events interface {
	Save(ctx context.Context, event sdktypes.Event) (sdktypes.EventID, error)
	Get(ctx context.Context, eventID sdktypes.EventID) (sdktypes.Event, error)
	// List returns events without their data.
	List(ctx context.Context, filter ListEventsFilter) ([]sdktypes.Event, error)
}
