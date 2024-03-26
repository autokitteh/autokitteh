// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: autokitteh/connections/v1/svc.proto

package connectionsv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1"
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
	// ConnectionsServiceName is the fully-qualified name of the ConnectionsService service.
	ConnectionsServiceName = "autokitteh.connections.v1.ConnectionsService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ConnectionsServiceCreateProcedure is the fully-qualified name of the ConnectionsService's Create
	// RPC.
	ConnectionsServiceCreateProcedure = "/autokitteh.connections.v1.ConnectionsService/Create"
	// ConnectionsServiceDeleteProcedure is the fully-qualified name of the ConnectionsService's Delete
	// RPC.
	ConnectionsServiceDeleteProcedure = "/autokitteh.connections.v1.ConnectionsService/Delete"
	// ConnectionsServiceUpdateProcedure is the fully-qualified name of the ConnectionsService's Update
	// RPC.
	ConnectionsServiceUpdateProcedure = "/autokitteh.connections.v1.ConnectionsService/Update"
	// ConnectionsServiceGetProcedure is the fully-qualified name of the ConnectionsService's Get RPC.
	ConnectionsServiceGetProcedure = "/autokitteh.connections.v1.ConnectionsService/Get"
	// ConnectionsServiceListProcedure is the fully-qualified name of the ConnectionsService's List RPC.
	ConnectionsServiceListProcedure = "/autokitteh.connections.v1.ConnectionsService/List"
	// ConnectionsServiceDelete1Procedure is the fully-qualified name of the ConnectionsService's
	// Delete1 RPC.
	ConnectionsServiceDelete1Procedure = "/autokitteh.connections.v1.ConnectionsService/Delete1"
)

// ConnectionsServiceClient is a client for the autokitteh.connections.v1.ConnectionsService
// service.
type ConnectionsServiceClient interface {
	// Initiated indirectly by an autokitteh user, based on an registered integration.
	Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error)
	Delete(context.Context, *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error)
	Update(context.Context, *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error)
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
	Delete1(context.Context, *connect.Request[v1.Delete1Request]) (*connect.Response[v1.Delete1Response], error)
}

// NewConnectionsServiceClient constructs a client for the
// autokitteh.connections.v1.ConnectionsService service. By default, it uses the Connect protocol
// with the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed requests. To
// use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or connect.WithGRPCWeb()
// options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewConnectionsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ConnectionsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &connectionsServiceClient{
		create: connect.NewClient[v1.CreateRequest, v1.CreateResponse](
			httpClient,
			baseURL+ConnectionsServiceCreateProcedure,
			opts...,
		),
		delete: connect.NewClient[v1.DeleteRequest, v1.DeleteResponse](
			httpClient,
			baseURL+ConnectionsServiceDeleteProcedure,
			opts...,
		),
		update: connect.NewClient[v1.UpdateRequest, v1.UpdateResponse](
			httpClient,
			baseURL+ConnectionsServiceUpdateProcedure,
			opts...,
		),
		get: connect.NewClient[v1.GetRequest, v1.GetResponse](
			httpClient,
			baseURL+ConnectionsServiceGetProcedure,
			opts...,
		),
		list: connect.NewClient[v1.ListRequest, v1.ListResponse](
			httpClient,
			baseURL+ConnectionsServiceListProcedure,
			opts...,
		),
		delete1: connect.NewClient[v1.Delete1Request, v1.Delete1Response](
			httpClient,
			baseURL+ConnectionsServiceDelete1Procedure,
			opts...,
		),
	}
}

// connectionsServiceClient implements ConnectionsServiceClient.
type connectionsServiceClient struct {
	create  *connect.Client[v1.CreateRequest, v1.CreateResponse]
	delete  *connect.Client[v1.DeleteRequest, v1.DeleteResponse]
	update  *connect.Client[v1.UpdateRequest, v1.UpdateResponse]
	get     *connect.Client[v1.GetRequest, v1.GetResponse]
	list    *connect.Client[v1.ListRequest, v1.ListResponse]
	delete1 *connect.Client[v1.Delete1Request, v1.Delete1Response]
}

// Create calls autokitteh.connections.v1.ConnectionsService.Create.
func (c *connectionsServiceClient) Create(ctx context.Context, req *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error) {
	return c.create.CallUnary(ctx, req)
}

// Delete calls autokitteh.connections.v1.ConnectionsService.Delete.
func (c *connectionsServiceClient) Delete(ctx context.Context, req *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error) {
	return c.delete.CallUnary(ctx, req)
}

// Update calls autokitteh.connections.v1.ConnectionsService.Update.
func (c *connectionsServiceClient) Update(ctx context.Context, req *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error) {
	return c.update.CallUnary(ctx, req)
}

// Get calls autokitteh.connections.v1.ConnectionsService.Get.
func (c *connectionsServiceClient) Get(ctx context.Context, req *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// List calls autokitteh.connections.v1.ConnectionsService.List.
func (c *connectionsServiceClient) List(ctx context.Context, req *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return c.list.CallUnary(ctx, req)
}

// Delete1 calls autokitteh.connections.v1.ConnectionsService.Delete1.
func (c *connectionsServiceClient) Delete1(ctx context.Context, req *connect.Request[v1.Delete1Request]) (*connect.Response[v1.Delete1Response], error) {
	return c.delete1.CallUnary(ctx, req)
}

// ConnectionsServiceHandler is an implementation of the
// autokitteh.connections.v1.ConnectionsService service.
type ConnectionsServiceHandler interface {
	// Initiated indirectly by an autokitteh user, based on an registered integration.
	Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error)
	Delete(context.Context, *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error)
	Update(context.Context, *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error)
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
	Delete1(context.Context, *connect.Request[v1.Delete1Request]) (*connect.Response[v1.Delete1Response], error)
}

// NewConnectionsServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewConnectionsServiceHandler(svc ConnectionsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	connectionsServiceCreateHandler := connect.NewUnaryHandler(
		ConnectionsServiceCreateProcedure,
		svc.Create,
		opts...,
	)
	connectionsServiceDeleteHandler := connect.NewUnaryHandler(
		ConnectionsServiceDeleteProcedure,
		svc.Delete,
		opts...,
	)
	connectionsServiceUpdateHandler := connect.NewUnaryHandler(
		ConnectionsServiceUpdateProcedure,
		svc.Update,
		opts...,
	)
	connectionsServiceGetHandler := connect.NewUnaryHandler(
		ConnectionsServiceGetProcedure,
		svc.Get,
		opts...,
	)
	connectionsServiceListHandler := connect.NewUnaryHandler(
		ConnectionsServiceListProcedure,
		svc.List,
		opts...,
	)
	connectionsServiceDelete1Handler := connect.NewUnaryHandler(
		ConnectionsServiceDelete1Procedure,
		svc.Delete1,
		opts...,
	)
	return "/autokitteh.connections.v1.ConnectionsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ConnectionsServiceCreateProcedure:
			connectionsServiceCreateHandler.ServeHTTP(w, r)
		case ConnectionsServiceDeleteProcedure:
			connectionsServiceDeleteHandler.ServeHTTP(w, r)
		case ConnectionsServiceUpdateProcedure:
			connectionsServiceUpdateHandler.ServeHTTP(w, r)
		case ConnectionsServiceGetProcedure:
			connectionsServiceGetHandler.ServeHTTP(w, r)
		case ConnectionsServiceListProcedure:
			connectionsServiceListHandler.ServeHTTP(w, r)
		case ConnectionsServiceDelete1Procedure:
			connectionsServiceDelete1Handler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedConnectionsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedConnectionsServiceHandler struct{}

func (UnimplementedConnectionsServiceHandler) Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.connections.v1.ConnectionsService.Create is not implemented"))
}

func (UnimplementedConnectionsServiceHandler) Delete(context.Context, *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.connections.v1.ConnectionsService.Delete is not implemented"))
}

func (UnimplementedConnectionsServiceHandler) Update(context.Context, *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.connections.v1.ConnectionsService.Update is not implemented"))
}

func (UnimplementedConnectionsServiceHandler) Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.connections.v1.ConnectionsService.Get is not implemented"))
}

func (UnimplementedConnectionsServiceHandler) List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.connections.v1.ConnectionsService.List is not implemented"))
}

func (UnimplementedConnectionsServiceHandler) Delete1(context.Context, *connect.Request[v1.Delete1Request]) (*connect.Response[v1.Delete1Response], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.connections.v1.ConnectionsService.Delete1 is not implemented"))
}
