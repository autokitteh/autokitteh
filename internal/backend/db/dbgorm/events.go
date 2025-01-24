package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) saveEvent(ctx context.Context, event *scheme.Event) error {
	return gdb.wdb.WithContext(ctx).Create(event).Error
}

func (gdb *gormdb) deleteEvent(ctx context.Context, eventID uuid.UUID) error {
	return gdb.wdb.WithContext(ctx).Delete(&scheme.Event{}, "event_id = ?", eventID).Error // NOTE: eventID isn't a primary key
}

func (gdb *gormdb) getEvent(ctx context.Context, eventID uuid.UUID) (*scheme.Event, error) {
	return getOne[scheme.Event](gdb.rdb.WithContext(ctx), "event_id = ?", eventID)
}

func (gdb *gormdb) listEvents(ctx context.Context, filter sdkservices.ListEventsFilter) ([]scheme.Event, error) {
	q := gdb.rdb.WithContext(ctx)

	q = withProjectOrgID(q, filter.OrgID, "events")

	if filter.ProjectID.IsValid() {
		q = q.Where("project_id = ?", filter.ProjectID.UUIDValue())
	}
	if filter.IntegrationID.IsValid() {
		q = q.Where("integration_id = ?", filter.IntegrationID.UUIDValue())
	}
	if filter.DestinationID.IsValid() {
		q = q.Where("destination_id = ?", filter.DestinationID.UUIDValue())
	}
	if filter.EventType != "" {
		q = q.Where("event_type = ?", filter.EventType)
	}
	if filter.CreatedAfter != nil {
		q = q.Where("created_at > ?", filter.CreatedAfter)
	}

	if filter.Limit != 0 {
		q = q.Limit(filter.Limit)
	}

	q = q.Where("seq >= ?", filter.MinSequenceNumber)

	if filter.Order == sdkservices.ListOrderAscending {
		q = q.Order("seq asc")
	} else {
		q = q.Order("seq desc") // default to desc
	}

	var es []scheme.Event
	if err := q.Omit("data").Find(&es).Error; err != nil {
		return nil, err
	}
	return es, nil
}

func (db *gormdb) SaveEvent(ctx context.Context, event sdktypes.Event) error {
	if err := event.Strict(); err != nil {
		return err
	}

	connectionID := event.DestinationID().ToConnectionID()

	pid, err := db.GetProjectIDOf(ctx, event.DestinationID())
	if err != nil {
		return fmt.Errorf("get project id: %w", err)
	}

	e := scheme.Event{
		Base:          based(ctx),
		ProjectID:     pid.UUIDValue(),
		EventID:       event.ID().UUIDValue(),
		DestinationID: event.DestinationID().UUIDValue(),
		ConnectionID:  uuidPtrOrNil(connectionID),
		TriggerID:     uuidPtrOrNil(event.DestinationID().ToTriggerID()),
		EventType:     event.Type(),
		Data:          kittehs.Must1(json.Marshal(event.Data())),
		Memo:          kittehs.Must1(json.Marshal(event.Memo())),
	}

	if connectionID.IsValid() { // only if exists
		conn, err := db.GetConnection(ctx, connectionID)
		if err != nil {
			return fmt.Errorf("get connection: %w", err)
		}

		if !conn.IsValid() {
			return sdkerrors.NewInvalidArgumentError("invalid event connection")
		}

		integrationID := conn.IntegrationID().UUIDValue()
		e.IntegrationID = &integrationID
	}

	return translateError(db.saveEvent(ctx, &e))
}

func (db *gormdb) GetEventByID(ctx context.Context, eventID sdktypes.EventID) (sdktypes.Event, error) {
	e, err := db.getEvent(ctx, eventID.UUIDValue())
	if e == nil || err != nil {
		return sdktypes.InvalidEvent, translateError(err)
	}
	return scheme.ParseEvent(*e)
}

func (db *gormdb) ListEvents(ctx context.Context, filter sdkservices.ListEventsFilter) ([]sdktypes.Event, error) {
	events, err := db.listEvents(ctx, filter)
	if events == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(events, scheme.ParseEvent)
}

func (db *gormdb) GetLatestEventSequence(ctx context.Context) (uint64, error) {
	// NOTE: called from workflow, not protected by user context
	var s scheme.Event
	if err := db.rdb.WithContext(ctx).Last(&s).Error; err != nil {
		return 0, translateError(err)
	}
	return s.Seq, nil
}
