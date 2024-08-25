package remotert

import (
	"net"

	"go.autokitteh.dev/autokitteh/runtimes/remotert/pb"
	"google.golang.org/grpc"
)

type workerServer struct {
	pb.UnimplementedWorkerServer
}

var ws = workerServer{}

func RegisterWorkerGRPCEndpoints(ln net.Listener) error {
	// muxes.NoAuth.Handle("/")
	srv := grpc.NewServer()
	pb.RegisterWorkerServer(srv, &ws)
	return srv.Serve(ln)
}
