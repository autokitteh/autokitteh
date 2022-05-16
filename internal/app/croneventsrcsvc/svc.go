package croneventsrcsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/robfig/cron"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/pkg/kvstore"

	pb "github.com/autokitteh/autokitteh/gen/proto/stubs/go/croneventsrc"

	L "github.com/autokitteh/autokitteh/pkg/l"
)

var EventTypes = []string{"tick"}

type Config struct {
	EventSourceID apieventsrc.EventSourceID `envconfig:"EVENT_SOURCE_ID" json:"event_source_id"`

	// This might be wanted off in real deployment - trigger can be done by external HTTP source
	// so only one cron svc will be active at any given tick. If this is != 0, all instances
	// will try to schedule which might trigger unwanted redundant ticks.
	LocalTickInterval           time.Duration `envconfig:"LOCAL_TICK_INTERVAL" default:"1m" json:"local_tick_interval"`
	LocalTickIntervalOffsetRand bool          `envconfig:"LOCAL_TICK_INTERVAL_RAND_OFFSET" default:"true" json:"local_tick_interval_rand_offset"`
}

type Svc struct {
	pb.UnimplementedCronEventSourceServer

	Config       Config
	StateStore   kvstore.Store
	Events       *events.Events
	EventSources eventsrcsstore.Store
	L            L.Nullable
}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	if srv != nil {
		pb.RegisterCronEventSourceServer(srv, s)
	}

	if gw != nil {
		if err := pb.RegisterCronEventSourceHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) Start() {
	if ival := s.Config.LocalTickInterval; ival != 0 {
		s.L.Info("local cron ticker active", "interval", ival)

		go func() {
			if s.Config.LocalTickIntervalOffsetRand {
				d := time.Duration(rand.Int63n(int64(ival)))
				s.L.Debug("delaying first tick with random duration", "d", d)
				time.Sleep(d)
			}

			s.L.Debug("first local trigger tick")

			// trigger first tick
			go s.tick()

			ticker := time.Tick(ival)
			for range ticker {
				s.L.Debug("local trigger tick")
				go s.tick()
			}
		}()
	}
}

func (s *Svc) getLast(ctx context.Context, pid apiproject.ProjectID, name string) (time.Time, error) {
	bs, err := s.StateStore.Get(ctx, fmt.Sprintf("%v/%s", pid, name))
	if err != nil {
		if errors.Is(err, kvstore.ErrNotFound) {
			return time.Time{}, nil
		}

		return time.Time{}, fmt.Errorf("get last: %w", err)
	}

	var t time.Time
	if err := json.Unmarshal(bs, &t); err != nil {
		return time.Time{}, fmt.Errorf("unmarshal: %w", err)
	}

	return t, nil
}

func (s *Svc) setLast(ctx context.Context, pid apiproject.ProjectID, name string, t time.Time) error {
	bs, _ := json.Marshal(t)
	if err := s.StateStore.Put(ctx, fmt.Sprintf("%v/%s", pid, name), bs); err != nil {
		return fmt.Errorf("set last: %w", err)
	}
	return nil
}

func (s *Svc) Tick(context.Context, *pb.TickRequest) (*pb.TickResponse, error) {
	go s.tick()

	return &pb.TickResponse{}, nil
}

func (s *Svc) tick() {
	ctx := context.Background()

	s.L.Debug("tick")

	// TODO: this might get very big. scale.
	bs, err := s.EventSources.GetProjectBindings(ctx, &s.Config.EventSourceID, nil, "", "", true)
	if err != nil {
		s.L.Error("get bindings error", "err", err)
		return
	}

	// TODO: [# spec-assoc-todo #] we coould use spec as the association token
	// so all bindings with the same spec will receive the same event and the
	// mux will be done by ak itself.
	for _, b := range bs {
		now := time.Now()

		l := s.L.With("name", b.Name(), "spec", b.SourceConfig(), "now", now)

		sched, err := cron.ParseStandard(b.SourceConfig())
		if err != nil {
			l.Error("invalid source config", "err", err)
			continue
		}

		since, err := s.getLast(ctx, b.ProjectID(), b.Name())
		if err != nil {
			l.Error("get last error", "err", err)
			continue
		}

		if since.IsZero() {
			l.Debug("never ran, setting since to now")
			since = time.Now()

			if err := s.setLast(ctx, b.ProjectID(), b.Name(), now); err != nil {
				l.Error("set last error", "err", err)
			}

			continue
		}

		l = l.With("since", since)

		next := sched.Next(since)
		trigger := !next.IsZero() && next.Before(now)

		if !trigger && !next.IsZero() {
			l = l.With("due", next.Sub(now))
		}

		l.Debug("next", "next", next, "trigger", trigger)

		if trigger {
			if err := s.setLast(ctx, b.ProjectID(), b.Name(), now); err != nil {
				l.Error("set last error", "err", err)
				continue
			}

			l.Debug("sending event")

			id, err := s.Events.IngestEvent(
				ctx,
				s.Config.EventSourceID,
				/* assoc */ "",
				/* originalID */ now.String(),
				/* type */ "tick",
				map[string]*apivalues.Value{
					"t":    apivalues.Time(now),
					"prev": apivalues.Time(since), // approx since it's  global.
				},
				nil,
			)

			if err != nil {
				l.Error("ingest event error", "err", err)
				continue
			}

			l.Debug("generated event", "id", id)
		}
	}
}

var ErrInvalidCronspec = errors.New("invalid cronspec")

func (s *Svc) Add(ctx context.Context, pid apiproject.ProjectID, name, spec string) error {
	// TODO: validate cronspec
	if _, err := cron.ParseStandard(spec); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidCronspec, err.Error())
	}

	if err := s.EventSources.AddProjectBinding(
		ctx,
		s.Config.EventSourceID,
		pid,
		name,
		"", // [# ./spec-assoc-todo #]
		spec,
		true,
		(&apieventsrc.EventSourceProjectBindingSettings{}).SetEnabled(true),
	); err != nil {
		return fmt.Errorf("bind: %w", err)
	}

	return nil
}

func (s *Svc) Bind(ctx context.Context, req *pb.BindRequest) (*pb.BindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	if err := s.Add(ctx, apiproject.ProjectID(req.ProjectId), req.Name, req.Cronspec); err != nil {
		if errors.Is(err, ErrInvalidCronspec) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}

		return nil, status.Errorf(codes.InvalidArgument, "add: %v", err)
	}

	return &pb.BindResponse{}, nil
}

func (s *Svc) Unbind(ctx context.Context, req *pb.UnbindRequest) (*pb.UnbindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	// TODO

	return &pb.UnbindResponse{}, nil
}
