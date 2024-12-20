package events

import (
	"context"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
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
	if err := authz.CheckContext(ctx, id, "read:get"); err != nil {
		return sdktypes.InvalidEvent, err
	}

	return e.db.GetEventByID(ctx, id)
}

func (e *events) List(ctx context.Context, filter sdkservices.ListEventsFilter) ([]sdktypes.Event, error) {
	if !filter.OrgID.IsValid() && !filter.ProjectID.IsValid() && !filter.DestinationID.IsValid() {
		filter.OrgID = authcontext.GetAuthnInferredOrgID(ctx)
	}

	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidEventID,
		"read:list",
		authz.WithData("filter", filter),
		authz.WithAssociationWithID("destination", filter.DestinationID),
		authz.WithAssociationWithID("project", filter.ProjectID),
		authz.WithAssociationWithID("org", filter.OrgID),
	); err != nil {
		return nil, err
	}

	return e.db.ListEvents(ctx, filter)
}

func (e *events) Save(ctx context.Context, event sdktypes.Event) (sdktypes.EventID, error) {
	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidEventID,
		"create:save",
		authz.WithData("event", event),
		authz.WithAssociationWithID("destination", event.DestinationID().AsID()),
	); err != nil {
		return sdktypes.InvalidEventID, err
	}

	event = event.WithNewID()
	if event.CreatedAt() == time.Unix(0, 0).UTC() {
		event = event.WithCreatedAt(time.Now())
	}

	if err := e.db.SaveEvent(ctx, event); err != nil {
		return sdktypes.InvalidEventID, err
	}

	return event.ID(), nil
}
