package sdkservices

import (
	"context"
	"time"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListEventsFilter struct {
	OrgID             sdktypes.OrgID
	ProjectID         sdktypes.ProjectID
	IntegrationID     sdktypes.IntegrationID
	DestinationID     sdktypes.EventDestinationID
	EventType         string
	Limit             int
	CreatedAfter      *time.Time
	MinSequenceNumber uint64
	Order             ListOrder
}

func (f ListEventsFilter) AnyIDSpecified() bool {
	// Do not put IntegrationID here - it does not limit the scope of the results org-wise
	return f.OrgID.IsValid() || f.ProjectID.IsValid() || f.DestinationID.IsValid()
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
