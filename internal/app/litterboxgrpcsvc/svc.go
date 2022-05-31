package litterboxgrpcsvc

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pbsvc.RegisterLitterBoxServer(srv, s)

	if gw != nil {
		if err := pbsvc.RegisterLitterBoxHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) Setup(ctx context.Context, req *pbsvc.SetupRequest) (*pbsvc.SetupResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	id, err := s.LitterBox.Setup(ctx, req.Name, req.Source)
	if err != nil {
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

func (s *Svc) Run(req *pbsvc.RunRequest, srv pbsvc.LitterBox_RunServer) error {
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

			if s := upd.ProjectState(); s != nil {
				if err := srv.Send(&pbsvc.RunUpdate{
					Id:    string(id),
					State: s.PB(),
				}); err != nil {
					l.Error("send update error", "err", err)
					return
				}
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
