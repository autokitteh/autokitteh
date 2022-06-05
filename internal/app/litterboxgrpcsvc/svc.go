package litterboxgrpcsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/tools/txtar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pbsvc "go.autokitteh.dev/idl/go/litterboxsvc"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apivalues"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/litterbox"
)

type Svc struct {
	pbsvc.UnimplementedLitterBoxServer

	L L.Nullable

	LitterBox litterbox.LitterBox
}

var _ pbsvc.LitterBoxServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux, port int) {
	pbsvc.RegisterLitterBoxServer(srv, s)

	if gw != nil {
		// Streaming GRPC Gateway does not work in-process. Need to do FromEndpoint.
		if err := pbsvc.RegisterLitterBoxHandlerFromEndpoint(
			ctx,
			gw,
			fmt.Sprintf("127.0.0.1:%d", port),
			[]grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithMaxMsgSize(1000000),
				grpc.WithDefaultCallOptions(
					grpc.MaxCallRecvMsgSize(1000000),
					grpc.MaxCallSendMsgSize(1000000),
				),
			},
		); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) Setup(ctx context.Context, req *pbsvc.SetupRequest) (*pbsvc.SetupResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	sources := req.SourcesMap

	if sources == nil {
		sources = make(map[string][]byte)
	}

	if alt := req.Sources; alt != nil {
		a := txtar.Parse(alt)

		for _, f := range a.Files {
			if req.MainSourceName == "" {
				req.MainSourceName = f.Name
			}

			sources[f.Name] = f.Data
		}

		if len(sources) == 0 {
			req.MainSourceName = "auto.kitteh"
			sources[req.MainSourceName] = alt
		}
	}

	id, err := s.LitterBox.Setup(ctx, litterbox.LitterBoxID(req.Id), sources, req.MainSourceName)
	if err != nil {
		if errors.Is(err, litterbox.ErrNoSources) || errors.Is(err, litterbox.ErrMainNotSpecified) {
			return nil, status.Errorf(codes.InvalidArgument, "setup: %v", err)
		}

		return nil, status.Errorf(codes.Unknown, "setup: %v", err)
	}

	return &pbsvc.SetupResponse{Id: string(id)}, nil
}

func (s *Svc) Scoop(ctx context.Context, req *pbsvc.ScoopRequest) (*pbsvc.ScoopResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	if err := s.LitterBox.Scoop(ctx, litterbox.LitterBoxID(req.Id)); err != nil {
		return nil, status.Errorf(codes.Unknown, "scoop: %v", err)
	}

	return &pbsvc.ScoopResponse{}, nil
}

func (s *Svc) Event(req *pbsvc.EventRequest, srv pbsvc.LitterBox_EventServer) error {
	if err := req.Validate(); err != nil {
		return status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	id := litterbox.LitterBoxID(req.Id)

	l := s.L.With("litterbox_id", id)

	data, err := apivalues.StringValueMapFromProto(req.Event.Data)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid data: %v", err)
	}

	ev := litterbox.LitterBoxEvent{
		SrcBinding: req.Event.SrcBinding,
		Type:       req.Event.Type,
		OriginalID: req.Event.OriginalId,
		Data:       data,
	}

	ch := make(chan *apievent.TrackIngestEventUpdate, 16)

	go func() {
		for upd := range ch {
			l.Debug("got update", "upd", upd)

			if err := srv.Send(upd.PB()); err != nil {
				l.Error("send update error", "err", err)
				return
			}
		}
	}()

	if err := s.LitterBox.RunEvent(
		srv.Context(),
		id,
		&ev,
		ch,
	); err != nil {
		return status.Errorf(codes.Unknown, "runevent: %v", err)
	}

	return nil
}
