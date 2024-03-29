// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: autokitteh/integrations/v1/svc.proto

package integrationsv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
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
	// IntegrationsServiceName is the fully-qualified name of the IntegrationsService service.
	IntegrationsServiceName = "autokitteh.integrations.v1.IntegrationsService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// IntegrationsServiceGetProcedure is the fully-qualified name of the IntegrationsService's Get RPC.
	IntegrationsServiceGetProcedure = "/autokitteh.integrations.v1.IntegrationsService/Get"
	// IntegrationsServiceListProcedure is the fully-qualified name of the IntegrationsService's List
	// RPC.
	IntegrationsServiceListProcedure = "/autokitteh.integrations.v1.IntegrationsService/List"
	// IntegrationsServiceConfigureProcedure is the fully-qualified name of the IntegrationsService's
	// Configure RPC.
	IntegrationsServiceConfigureProcedure = "/autokitteh.integrations.v1.IntegrationsService/Configure"
	// IntegrationsServiceCallProcedure is the fully-qualified name of the IntegrationsService's Call
	// RPC.
	IntegrationsServiceCallProcedure = "/autokitteh.integrations.v1.IntegrationsService/Call"
)

// IntegrationsServiceClient is a client for the autokitteh.integrations.v1.IntegrationsService
// service.
type IntegrationsServiceClient interface {
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
	// Get all values for a specific configuration of the integration.
	// The returned values ExecutorIDs will be the integration id.
	Configure(context.Context, *connect.Request[v1.ConfigureRequest]) (*connect.Response[v1.ConfigureResponse], error)
	Call(context.Context, *connect.Request[v1.CallRequest]) (*connect.Response[v1.CallResponse], error)
}

// NewIntegrationsServiceClient constructs a client for the
// autokitteh.integrations.v1.IntegrationsService service. By default, it uses the Connect protocol
// with the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed requests. To
// use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or connect.WithGRPCWeb()
// options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewIntegrationsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) IntegrationsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &integrationsServiceClient{
		get: connect.NewClient[v1.GetRequest, v1.GetResponse](
			httpClient,
			baseURL+IntegrationsServiceGetProcedure,
			opts...,
		),
		list: connect.NewClient[v1.ListRequest, v1.ListResponse](
			httpClient,
			baseURL+IntegrationsServiceListProcedure,
			opts...,
		),
		configure: connect.NewClient[v1.ConfigureRequest, v1.ConfigureResponse](
			httpClient,
			baseURL+IntegrationsServiceConfigureProcedure,
			opts...,
		),
		call: connect.NewClient[v1.CallRequest, v1.CallResponse](
			httpClient,
			baseURL+IntegrationsServiceCallProcedure,
			opts...,
		),
	}
}

// integrationsServiceClient implements IntegrationsServiceClient.
type integrationsServiceClient struct {
	get       *connect.Client[v1.GetRequest, v1.GetResponse]
	list      *connect.Client[v1.ListRequest, v1.ListResponse]
	configure *connect.Client[v1.ConfigureRequest, v1.ConfigureResponse]
	call      *connect.Client[v1.CallRequest, v1.CallResponse]
}

// Get calls autokitteh.integrations.v1.IntegrationsService.Get.
func (c *integrationsServiceClient) Get(ctx context.Context, req *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// List calls autokitteh.integrations.v1.IntegrationsService.List.
func (c *integrationsServiceClient) List(ctx context.Context, req *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return c.list.CallUnary(ctx, req)
}

// Configure calls autokitteh.integrations.v1.IntegrationsService.Configure.
func (c *integrationsServiceClient) Configure(ctx context.Context, req *connect.Request[v1.ConfigureRequest]) (*connect.Response[v1.ConfigureResponse], error) {
	return c.configure.CallUnary(ctx, req)
}

// Call calls autokitteh.integrations.v1.IntegrationsService.Call.
func (c *integrationsServiceClient) Call(ctx context.Context, req *connect.Request[v1.CallRequest]) (*connect.Response[v1.CallResponse], error) {
	return c.call.CallUnary(ctx, req)
}

// IntegrationsServiceHandler is an implementation of the
// autokitteh.integrations.v1.IntegrationsService service.
type IntegrationsServiceHandler interface {
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
	// Get all values for a specific configuration of the integration.
	// The returned values ExecutorIDs will be the integration id.
	Configure(context.Context, *connect.Request[v1.ConfigureRequest]) (*connect.Response[v1.ConfigureResponse], error)
	Call(context.Context, *connect.Request[v1.CallRequest]) (*connect.Response[v1.CallResponse], error)
}

// NewIntegrationsServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewIntegrationsServiceHandler(svc IntegrationsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	integrationsServiceGetHandler := connect.NewUnaryHandler(
		IntegrationsServiceGetProcedure,
		svc.Get,
		opts...,
	)
	integrationsServiceListHandler := connect.NewUnaryHandler(
		IntegrationsServiceListProcedure,
		svc.List,
		opts...,
	)
	integrationsServiceConfigureHandler := connect.NewUnaryHandler(
		IntegrationsServiceConfigureProcedure,
		svc.Configure,
		opts...,
	)
	integrationsServiceCallHandler := connect.NewUnaryHandler(
		IntegrationsServiceCallProcedure,
		svc.Call,
		opts...,
	)
	return "/autokitteh.integrations.v1.IntegrationsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case IntegrationsServiceGetProcedure:
			integrationsServiceGetHandler.ServeHTTP(w, r)
		case IntegrationsServiceListProcedure:
			integrationsServiceListHandler.ServeHTTP(w, r)
		case IntegrationsServiceConfigureProcedure:
			integrationsServiceConfigureHandler.ServeHTTP(w, r)
		case IntegrationsServiceCallProcedure:
			integrationsServiceCallHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedIntegrationsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedIntegrationsServiceHandler struct{}

func (UnimplementedIntegrationsServiceHandler) Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.integrations.v1.IntegrationsService.Get is not implemented"))
}

func (UnimplementedIntegrationsServiceHandler) List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.integrations.v1.IntegrationsService.List is not implemented"))
}

func (UnimplementedIntegrationsServiceHandler) Configure(context.Context, *connect.Request[v1.ConfigureRequest]) (*connect.Response[v1.ConfigureResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.integrations.v1.IntegrationsService.Configure is not implemented"))
}

func (UnimplementedIntegrationsServiceHandler) Call(context.Context, *connect.Request[v1.CallRequest]) (*connect.Response[v1.CallResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.integrations.v1.IntegrationsService.Call is not implemented"))
}
