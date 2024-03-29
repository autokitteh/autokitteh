// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: autokitteh/secrets/v1/svc.proto

package secretsv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/secrets/v1"
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
	// SecretsServiceName is the fully-qualified name of the SecretsService service.
	SecretsServiceName = "autokitteh.secrets.v1.SecretsService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// SecretsServiceCreateProcedure is the fully-qualified name of the SecretsService's Create RPC.
	SecretsServiceCreateProcedure = "/autokitteh.secrets.v1.SecretsService/Create"
	// SecretsServiceGetProcedure is the fully-qualified name of the SecretsService's Get RPC.
	SecretsServiceGetProcedure = "/autokitteh.secrets.v1.SecretsService/Get"
	// SecretsServiceListProcedure is the fully-qualified name of the SecretsService's List RPC.
	SecretsServiceListProcedure = "/autokitteh.secrets.v1.SecretsService/List"
)

// SecretsServiceClient is a client for the autokitteh.secrets.v1.SecretsService service.
type SecretsServiceClient interface {
	// Create generates a new token to represent a connection's specified
	// key-value data, and associates them bidirectionally. If the same
	// request is sent N times, this method returns N different tokens.
	Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error)
	// Get retrieves a connection's key-value data based on the given token.
	// If the token isn’t found then we return an error.
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	// List enumerates all the tokens (0 or more) that are associated with a given
	// connection identifier. This enables autokitteh to dispatch/fan-out asynchronous
	// events that it receives from integrations through all the relevant connections.
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
}

// NewSecretsServiceClient constructs a client for the autokitteh.secrets.v1.SecretsService service.
// By default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped
// responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewSecretsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) SecretsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &secretsServiceClient{
		create: connect.NewClient[v1.CreateRequest, v1.CreateResponse](
			httpClient,
			baseURL+SecretsServiceCreateProcedure,
			opts...,
		),
		get: connect.NewClient[v1.GetRequest, v1.GetResponse](
			httpClient,
			baseURL+SecretsServiceGetProcedure,
			opts...,
		),
		list: connect.NewClient[v1.ListRequest, v1.ListResponse](
			httpClient,
			baseURL+SecretsServiceListProcedure,
			opts...,
		),
	}
}

// secretsServiceClient implements SecretsServiceClient.
type secretsServiceClient struct {
	create *connect.Client[v1.CreateRequest, v1.CreateResponse]
	get    *connect.Client[v1.GetRequest, v1.GetResponse]
	list   *connect.Client[v1.ListRequest, v1.ListResponse]
}

// Create calls autokitteh.secrets.v1.SecretsService.Create.
func (c *secretsServiceClient) Create(ctx context.Context, req *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error) {
	return c.create.CallUnary(ctx, req)
}

// Get calls autokitteh.secrets.v1.SecretsService.Get.
func (c *secretsServiceClient) Get(ctx context.Context, req *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// List calls autokitteh.secrets.v1.SecretsService.List.
func (c *secretsServiceClient) List(ctx context.Context, req *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return c.list.CallUnary(ctx, req)
}

// SecretsServiceHandler is an implementation of the autokitteh.secrets.v1.SecretsService service.
type SecretsServiceHandler interface {
	// Create generates a new token to represent a connection's specified
	// key-value data, and associates them bidirectionally. If the same
	// request is sent N times, this method returns N different tokens.
	Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error)
	// Get retrieves a connection's key-value data based on the given token.
	// If the token isn’t found then we return an error.
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	// List enumerates all the tokens (0 or more) that are associated with a given
	// connection identifier. This enables autokitteh to dispatch/fan-out asynchronous
	// events that it receives from integrations through all the relevant connections.
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
}

// NewSecretsServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewSecretsServiceHandler(svc SecretsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	secretsServiceCreateHandler := connect.NewUnaryHandler(
		SecretsServiceCreateProcedure,
		svc.Create,
		opts...,
	)
	secretsServiceGetHandler := connect.NewUnaryHandler(
		SecretsServiceGetProcedure,
		svc.Get,
		opts...,
	)
	secretsServiceListHandler := connect.NewUnaryHandler(
		SecretsServiceListProcedure,
		svc.List,
		opts...,
	)
	return "/autokitteh.secrets.v1.SecretsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case SecretsServiceCreateProcedure:
			secretsServiceCreateHandler.ServeHTTP(w, r)
		case SecretsServiceGetProcedure:
			secretsServiceGetHandler.ServeHTTP(w, r)
		case SecretsServiceListProcedure:
			secretsServiceListHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedSecretsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedSecretsServiceHandler struct{}

func (UnimplementedSecretsServiceHandler) Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.secrets.v1.SecretsService.Create is not implemented"))
}

func (UnimplementedSecretsServiceHandler) Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.secrets.v1.SecretsService.Get is not implemented"))
}

func (UnimplementedSecretsServiceHandler) List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.secrets.v1.SecretsService.List is not implemented"))
}
