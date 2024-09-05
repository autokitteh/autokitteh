package remotert

import (
	"context"
	"fmt"
	"sync"

	"go.autokitteh.dev/autokitteh/runtimes/remotert/pb"
	"google.golang.org/grpc"
)

type workerServer struct {
	pb.UnimplementedWorkerServer
	runnerIDsToRuntime map[string]*svc
	mu                 *sync.Mutex
}

var ws = workerServer{
	runnerIDsToRuntime: map[string]*svc{},
	mu:                 new(sync.Mutex),
}

func (w *workerServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	var resp pb.HealthResponse

	// TODO: Health check

	return &resp, nil
}

func (w *workerServer) Print(ctx context.Context, req *pb.PrintRequest) (*pb.PrintResponse, error) {
	var resp pb.PrintResponse

	return &resp, nil
}

func (w *workerServer) Done(ctx context.Context, req *pb.DoneRequest) (*pb.DoneResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	resp := &pb.DoneResponse{}
	if !ok {
		return resp, nil
	}

	if req.Error != "" {
		runner.errorChan <- req.Error
		return resp, nil
	}
	go func() {
		fmt.Println("EFI--- Done", req.RunnerId)
		runner.doneChan <- req.Result
	}()

	return resp, nil
}

func (w *workerServer) Activity(ctx context.Context, req *pb.ActivityRequest) (*pb.ActivityResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.ActivityResponse{
			Error: "Unknown runner id",
		}, nil
	}

	go func() {
		fmt.Println("EFI---CALLING ACTIVITY", req.RunnerId)
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
