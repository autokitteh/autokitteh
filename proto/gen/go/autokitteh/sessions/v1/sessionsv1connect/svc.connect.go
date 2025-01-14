// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: autokitteh/sessions/v1/svc.proto

package sessionsv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion0_1_0

const (
	// SessionsServiceName is the fully-qualified name of the SessionsService service.
	SessionsServiceName = "autokitteh.sessions.v1.SessionsService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// SessionsServiceStartProcedure is the fully-qualified name of the SessionsService's Start RPC.
	SessionsServiceStartProcedure = "/autokitteh.sessions.v1.SessionsService/Start"
	// SessionsServiceStopProcedure is the fully-qualified name of the SessionsService's Stop RPC.
	SessionsServiceStopProcedure = "/autokitteh.sessions.v1.SessionsService/Stop"
	// SessionsServiceListProcedure is the fully-qualified name of the SessionsService's List RPC.
	SessionsServiceListProcedure = "/autokitteh.sessions.v1.SessionsService/List"
	// SessionsServiceGetProcedure is the fully-qualified name of the SessionsService's Get RPC.
	SessionsServiceGetProcedure = "/autokitteh.sessions.v1.SessionsService/Get"
	// SessionsServiceGetLogProcedure is the fully-qualified name of the SessionsService's GetLog RPC.
	SessionsServiceGetLogProcedure = "/autokitteh.sessions.v1.SessionsService/GetLog"
	// SessionsServiceDeleteProcedure is the fully-qualified name of the SessionsService's Delete RPC.
	SessionsServiceDeleteProcedure = "/autokitteh.sessions.v1.SessionsService/Delete"
)

// SessionsServiceClient is a client for the autokitteh.sessions.v1.SessionsService service.
type SessionsServiceClient interface {
	Start(context.Context, *connect.Request[v1.StartRequest]) (*connect.Response[v1.StartResponse], error)
	// Will always try first to gracefully terminate the session.
	// Blocks only if `force` and forceDelay > 0`.
	Stop(context.Context, *connect.Request[v1.StopRequest]) (*connect.Response[v1.StopResponse], error)
	// List returns events without their data.
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	GetLog(context.Context, *connect.Request[v1.GetLogRequest]) (*connect.Response[v1.GetLogResponse], error)
	Delete(context.Context, *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error)
}

// NewSessionsServiceClient constructs a client for the autokitteh.sessions.v1.SessionsService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewSessionsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) SessionsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &sessionsServiceClient{
		start: connect.NewClient[v1.StartRequest, v1.StartResponse](
			httpClient,
			baseURL+SessionsServiceStartProcedure,
			opts...,
		),
		stop: connect.NewClient[v1.StopRequest, v1.StopResponse](
			httpClient,
			baseURL+SessionsServiceStopProcedure,
			opts...,
		),
		list: connect.NewClient[v1.ListRequest, v1.ListResponse](
			httpClient,
			baseURL+SessionsServiceListProcedure,
			opts...,
		),
		get: connect.NewClient[v1.GetRequest, v1.GetResponse](
			httpClient,
			baseURL+SessionsServiceGetProcedure,
			opts...,
		),
		getLog: connect.NewClient[v1.GetLogRequest, v1.GetLogResponse](
			httpClient,
			baseURL+SessionsServiceGetLogProcedure,
			opts...,
		),
		delete: connect.NewClient[v1.DeleteRequest, v1.DeleteResponse](
			httpClient,
			baseURL+SessionsServiceDeleteProcedure,
			opts...,
		),
	}
}

// sessionsServiceClient implements SessionsServiceClient.
type sessionsServiceClient struct {
	start  *connect.Client[v1.StartRequest, v1.StartResponse]
	stop   *connect.Client[v1.StopRequest, v1.StopResponse]
	list   *connect.Client[v1.ListRequest, v1.ListResponse]
	get    *connect.Client[v1.GetRequest, v1.GetResponse]
	getLog *connect.Client[v1.GetLogRequest, v1.GetLogResponse]
	delete *connect.Client[v1.DeleteRequest, v1.DeleteResponse]
}

// Start calls autokitteh.sessions.v1.SessionsService.Start.
func (c *sessionsServiceClient) Start(ctx context.Context, req *connect.Request[v1.StartRequest]) (*connect.Response[v1.StartResponse], error) {
	return c.start.CallUnary(ctx, req)
}

// Stop calls autokitteh.sessions.v1.SessionsService.Stop.
func (c *sessionsServiceClient) Stop(ctx context.Context, req *connect.Request[v1.StopRequest]) (*connect.Response[v1.StopResponse], error) {
	return c.stop.CallUnary(ctx, req)
}

// List calls autokitteh.sessions.v1.SessionsService.List.
func (c *sessionsServiceClient) List(ctx context.Context, req *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return c.list.CallUnary(ctx, req)
}

// Get calls autokitteh.sessions.v1.SessionsService.Get.
func (c *sessionsServiceClient) Get(ctx context.Context, req *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// GetLog calls autokitteh.sessions.v1.SessionsService.GetLog.
func (c *sessionsServiceClient) GetLog(ctx context.Context, req *connect.Request[v1.GetLogRequest]) (*connect.Response[v1.GetLogResponse], error) {
	return c.getLog.CallUnary(ctx, req)
}

// Delete calls autokitteh.sessions.v1.SessionsService.Delete.
func (c *sessionsServiceClient) Delete(ctx context.Context, req *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error) {
	return c.delete.CallUnary(ctx, req)
}

// SessionsServiceHandler is an implementation of the autokitteh.sessions.v1.SessionsService
// service.
type SessionsServiceHandler interface {
	Start(context.Context, *connect.Request[v1.StartRequest]) (*connect.Response[v1.StartResponse], error)
	// Will always try first to gracefully terminate the session.
	// Blocks only if `force` and forceDelay > 0`.
	Stop(context.Context, *connect.Request[v1.StopRequest]) (*connect.Response[v1.StopResponse], error)
	// List returns events without their data.
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	GetLog(context.Context, *connect.Request[v1.GetLogRequest]) (*connect.Response[v1.GetLogResponse], error)
	Delete(context.Context, *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error)
}

// NewSessionsServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewSessionsServiceHandler(svc SessionsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	sessionsServiceStartHandler := connect.NewUnaryHandler(
		SessionsServiceStartProcedure,
		svc.Start,
		opts...,
	)
	sessionsServiceStopHandler := connect.NewUnaryHandler(
		SessionsServiceStopProcedure,
		svc.Stop,
		opts...,
	)
	sessionsServiceListHandler := connect.NewUnaryHandler(
		SessionsServiceListProcedure,
		svc.List,
		opts...,
	)
	sessionsServiceGetHandler := connect.NewUnaryHandler(
		SessionsServiceGetProcedure,
		svc.Get,
		opts...,
	)
	sessionsServiceGetLogHandler := connect.NewUnaryHandler(
		SessionsServiceGetLogProcedure,
		svc.GetLog,
		opts...,
	)
	sessionsServiceDeleteHandler := connect.NewUnaryHandler(
		SessionsServiceDeleteProcedure,
		svc.Delete,
		opts...,
	)
	return "/autokitteh.sessions.v1.SessionsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case SessionsServiceStartProcedure:
			sessionsServiceStartHandler.ServeHTTP(w, r)
		case SessionsServiceStopProcedure:
			sessionsServiceStopHandler.ServeHTTP(w, r)
		case SessionsServiceListProcedure:
			sessionsServiceListHandler.ServeHTTP(w, r)
		case SessionsServiceGetProcedure:
			sessionsServiceGetHandler.ServeHTTP(w, r)
		case SessionsServiceGetLogProcedure:
			sessionsServiceGetLogHandler.ServeHTTP(w, r)
		case SessionsServiceDeleteProcedure:
			sessionsServiceDeleteHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedSessionsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedSessionsServiceHandler struct{}

func (UnimplementedSessionsServiceHandler) Start(context.Context, *connect.Request[v1.StartRequest]) (*connect.Response[v1.StartResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.sessions.v1.SessionsService.Start is not implemented"))
}

func (UnimplementedSessionsServiceHandler) Stop(context.Context, *connect.Request[v1.StopRequest]) (*connect.Response[v1.StopResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.sessions.v1.SessionsService.Stop is not implemented"))
}

func (UnimplementedSessionsServiceHandler) List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.sessions.v1.SessionsService.List is not implemented"))
}

func (UnimplementedSessionsServiceHandler) Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.sessions.v1.SessionsService.Get is not implemented"))
}

func (UnimplementedSessionsServiceHandler) GetLog(context.Context, *connect.Request[v1.GetLogRequest]) (*connect.Response[v1.GetLogResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.sessions.v1.SessionsService.GetLog is not implemented"))
}

func (UnimplementedSessionsServiceHandler) Delete(context.Context, *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.sessions.v1.SessionsService.Delete is not implemented"))
}
