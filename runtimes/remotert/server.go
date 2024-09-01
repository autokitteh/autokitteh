package remotert

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/runtimes/remotert/pb"
	"google.golang.org/grpc"
)

type workerServer struct {
	pb.UnimplementedWorkerServer
	runnerIDsToRuntime map[string]*svc
}

var ws = workerServer{
	runnerIDsToRuntime: map[string]*svc{},
}

func (w *workerServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	var resp pb.HealthResponse

	// TODO: Health check

	return &resp, nil
}

func (w *workerServer) Done(ctx context.Context, req *pb.DoneRequest) (*pb.DoneResponse, error) {
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	resp := &pb.DoneResponse{}
	if !ok {
		return resp, nil
	}

	if req.Error != "" {
		runner.errorChan <- req.Error
		return resp, nil
	}

	runner.doneChan <- req.Result
	return resp, nil
}

func (w *workerServer) Activity(ctx context.Context, req *pb.ActivityRequest) (*pb.ActivityResponse, error) {
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	if !ok {
		return &pb.ActivityResponse{
			Error: "Unknown runner id",
		}, nil
	}

	go func() {
		fmt.Println("EFI---CALLING ACTIVITY")
		runner.runnerRequestsChan <- req
	}()

	return &pb.ActivityResponse{}, nil
}

var Server = func() *grpc.Server {
	srv := grpc.NewServer()
	pb.RegisterWorkerServer(srv, &ws)
	return srv
}()

// func Middleware(next http.Handler) http.HandlerFunc {

// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if strings.HasPrefix(r.URL.Path, "/Worker") {
// 			srv.ServeHTTP(w, r)
// 		} else {
// 			next.ServeHTTP(w, r)
// 		}
// 	})
// }
