package remotert

import (
	"context"

	"go.autokitteh.dev/autokitteh/runtimes/remotert/pb"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"google.golang.org/grpc"
)

type workerServer struct {
	pb.UnimplementedWorkerServer
	svcs map[string]*svc
}

var ws = workerServer{
	svcs: map[string]*svc{},
}

func (w *workerServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	var resp pb.HealthResponse

	// TODO: Health check

	return &resp, nil
}

func (w *workerServer) Activity(ctx context.Context, req *pb.ActivityRequest) (*pb.ActivityResponse, error) {
	runner, ok := w.svcs[req.RunnerId]
	if !ok {
		return &pb.ActivityResponse{
			Error: "Unknown runner id",
		}, nil
	}

	go func() {
		_, err := runner.cbs.Call(ctx, runner.runID.ToRunID(), sdktypes.NewStringValue(req.CallId), nil, nil)
		if err != nil {
			return
		}
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
