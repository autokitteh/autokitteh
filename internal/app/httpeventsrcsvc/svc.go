package httpeventsrcsvc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autokitteh/autokitteh/api/gen/stubs/go/httpeventsrc"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"github.com/autokitteh/autokitteh/internal/pkg/httpeventsrc"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type Svc struct {
	pb.UnimplementedHTTPEventSourceServer

	Src *httpeventsrc.HTTPEventSource
	L   L.Nullable
}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux, r *mux.Route) {
	r.HandlerFunc(s.handler)

	if srv != nil {
		pb.RegisterHTTPEventSourceServer(srv, s)
	}

	if gw != nil {
		if err := pb.RegisterHTTPEventSourceHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) Bind(ctx context.Context, req *pb.BindRequest) (*pb.BindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	if err := s.Src.Add(ctx, apiproject.ProjectID(req.ProjectId), req.Name, req.Routes); err != nil {
		return nil, status.Errorf(codes.Unknown, "create: %v", err)
	}

	return &pb.BindResponse{}, nil
}

func (s *Svc) Unbind(ctx context.Context, req *pb.UnbindRequest) (*pb.UnbindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	if err := s.Src.Remove(ctx, apiproject.ProjectID(req.ProjectId), req.Name); err != nil {
		return nil, status.Errorf(codes.Unknown, "remove: %v", err)
	}

	return &pb.UnbindResponse{}, nil
}

func (s *Svc) handler(w http.ResponseWriter, req *http.Request) {
	s.L.Debug("got request", "url", req.URL.String(), "method", req.Method)

	defer req.Body.Close()

	id, err := s.Src.Handle(req)
	if err != nil {
		if httpErr := (&httpeventsrc.HTTPError{}); errors.As(err, &httpErr) {
			httpErr.Write(w)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if id == "" {
		http.NotFound(w, req)
		return
	}

	w.Header().Add("X-AutoKitteh-Event-Id", string(id))
	w.WriteHeader(http.StatusCreated)

	// TODO: write a link to event ui frontend.

	resp := struct {
		EventID string `json:"event_id"`
	}{
		EventID: id.String(),
	}

	body, _ := json.Marshal(resp)
	_, _ = w.Write([]byte(body))
}
