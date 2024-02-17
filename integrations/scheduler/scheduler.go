package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

const (
	scope = "scheduler"

	// initInterval is the interval at which we check for new connections.
	initInterval = time.Second
)

type connection struct {
	schedule, timezone, memo string
	cronID                   cron.EntryID
}

type event struct {
	// Trigger settings.
	Schedule, Timezone, Memo string

	// Event instance.
	Timestamp  time.Time
	SinceEpoch int64
	Location   string

	Year, Month, Day, Weekday int
	Hour, Minute, Second      int
}

var (
	connections = map[string]connection{}
	cronTable   = cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
)

// detectNewConnections is a persistent goroutine that periodically
// checks for new connections and adds them to the cron table.
func detectNewConnections(l *zap.Logger, s sdkservices.Secrets, d sdkservices.Dispatcher) {
	ctx := context.Background()

	tokens, err := s.List(ctx, scope, "all")
	if err != nil {
		l.Error("Failed to list connections", zap.Error(err))
		return
	}
	for _, token := range tokens {
		// Ignore existing connections.
		if _, ok := connections[token]; ok {
			continue
		}
		// Add new connections to the cron table.
		// TODO(ENG-301): Support multiple distributed server instances.
		if conn, err := s.Get(ctx, scope, token); err == nil {
			c := connection{
				schedule: conn["schedule"],
				timezone: conn["timezone"],
				memo:     conn["memo"],
			}
			spec := c.schedule
			if c.timezone != "Local" {
				spec = fmt.Sprintf("CRON_TZ=%s %s", c.timezone, s)
			}
			id, err := cronTable.AddFunc(spec, dispatchEvents(ctx, l, d, token, c))
			if err != nil {
				l.Error("Failed to add cron schedule",
					zap.String("token", token),
					zap.Any("connection", c),
					zap.Error(err),
				)
				continue
			}
			c.cronID = id
			connections[token] = c
			cronTable.Start()
		}
	}
}

func dispatchEvents(ctx context.Context, l *zap.Logger, d sdkservices.Dispatcher, token string, conn connection) func() {
	return func() {
		now := time.Now()
		e := event{
			// Trigger settings.
			Schedule: conn.schedule,
			Timezone: conn.timezone,
			Memo:     conn.memo,

			// Event instance.
			Timestamp:  now,
			SinceEpoch: now.Unix(),
			Location:   now.Location().String(),

			Year:    now.Year(),
			Month:   int(now.Month()),
			Day:     now.Day(),
			Weekday: int(now.Weekday()),

			Hour:   now.Hour(),
			Minute: now.Minute(),
			Second: now.Second(),
		}

		wrapped, err := sdkvalues.DefaultValueWrapper.Wrap(e)
		if err != nil {
			l.Error("Failed to wrap cron event",
				zap.Any("event", e),
				zap.Error(err),
			)
			return
		}
		data, err := sdktypes.ValueToStringValuesMap(wrapped)
		if err != nil {
			l.Error("Failed to convert wrapped cron event",
				zap.Any("event", e),
				zap.Error(err),
			)
			return
		}
		proto := &sdktypes.EventPB{
			IntegrationId:    integrationID.String(),
			IntegrationToken: token,
			OriginalEventId:  strconv.FormatInt(now.UnixNano(), 10),
			EventType:        "cron_trigger",
			Data:             kittehs.TransformMapValues(data, sdktypes.ToProto),
		}
		event := kittehs.Must1(sdktypes.EventFromProto(proto))

		eventID, err := d.Dispatch(ctx, event, nil)
		if err != nil {
			l.Error("Dispatch failed",
				zap.String("connectionToken", token),
				zap.Error(err),
			)
			return
		}
		l.Debug("Dispatched",
			zap.String("connectionToken", token),
			zap.String("eventID", eventID.String()),
		)
	}
}
