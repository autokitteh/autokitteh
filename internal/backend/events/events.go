package events

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type events struct {
	z  *zap.Logger
	db db.DB
}

func New(z *zap.Logger, db db.DB) sdkservices.Events {
	return &events{db: db, z: z}
}

func (e *events) Get(ctx context.Context, id sdktypes.EventID) (sdktypes.Event, error) {
	return e.db.GetEventByID(ctx, id)
}

func (e *events) List(ctx context.Context, filter sdkservices.ListEventsFilter) ([]sdktypes.Event, error) {
	return e.db.ListEvents(ctx, filter)
}

func (e *events) Save(ctx context.Context, event sdktypes.Event) (sdktypes.EventID, error) {
	event = event.WithNewID()
	if event.CreatedAt() == time.Unix(0, 0).UTC() {
		event = event.WithCreatedAt(time.Now())
	}

	if err := e.db.SaveEvent(ctx, event); err != nil {
		return sdktypes.InvalidEventID, err
	}

	return event.ID(), nil
}
