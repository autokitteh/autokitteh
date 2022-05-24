package litterboxgrpcsvc

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pbsvc "go.autokitteh.dev/idl/go/litterboxsvc"

	"github.com/autokitteh/L"
)

type Svc struct {
	pbsvc.UnimplementedLitterBoxServer

	L L.Nullable
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
	return nil, nil
}

func (s *Svc) Run(req *pbsvc.RunRequest, srv pbsvc.LitterBox_RunServer) error {
	return nil
}

func (s *Svc) Scoop(ctx context.Context, req *pbsvc.ScoopRequest) (*pbsvc.ScoopResponse, error) {
	return nil, nil
}
