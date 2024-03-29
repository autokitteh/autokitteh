// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: autokitteh/envs/v1/svc.proto

package envsv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1"
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
	// EnvsServiceName is the fully-qualified name of the EnvsService service.
	EnvsServiceName = "autokitteh.envs.v1.EnvsService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// EnvsServiceListProcedure is the fully-qualified name of the EnvsService's List RPC.
	EnvsServiceListProcedure = "/autokitteh.envs.v1.EnvsService/List"
	// EnvsServiceCreateProcedure is the fully-qualified name of the EnvsService's Create RPC.
	EnvsServiceCreateProcedure = "/autokitteh.envs.v1.EnvsService/Create"
	// EnvsServiceGetProcedure is the fully-qualified name of the EnvsService's Get RPC.
	EnvsServiceGetProcedure = "/autokitteh.envs.v1.EnvsService/Get"
	// EnvsServiceRemoveProcedure is the fully-qualified name of the EnvsService's Remove RPC.
	EnvsServiceRemoveProcedure = "/autokitteh.envs.v1.EnvsService/Remove"
	// EnvsServiceUpdateProcedure is the fully-qualified name of the EnvsService's Update RPC.
	EnvsServiceUpdateProcedure = "/autokitteh.envs.v1.EnvsService/Update"
	// EnvsServiceSetVarProcedure is the fully-qualified name of the EnvsService's SetVar RPC.
	EnvsServiceSetVarProcedure = "/autokitteh.envs.v1.EnvsService/SetVar"
	// EnvsServiceRemoveVarProcedure is the fully-qualified name of the EnvsService's RemoveVar RPC.
	EnvsServiceRemoveVarProcedure = "/autokitteh.envs.v1.EnvsService/RemoveVar"
	// EnvsServiceGetVarsProcedure is the fully-qualified name of the EnvsService's GetVars RPC.
	EnvsServiceGetVarsProcedure = "/autokitteh.envs.v1.EnvsService/GetVars"
	// EnvsServiceRevealVarProcedure is the fully-qualified name of the EnvsService's RevealVar RPC.
	EnvsServiceRevealVarProcedure = "/autokitteh.envs.v1.EnvsService/RevealVar"
)

// EnvsServiceClient is a client for the autokitteh.envs.v1.EnvsService service.
type EnvsServiceClient interface {
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
	Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error)
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	Remove(context.Context, *connect.Request[v1.RemoveRequest]) (*connect.Response[v1.RemoveResponse], error)
	Update(context.Context, *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error)
	SetVar(context.Context, *connect.Request[v1.SetVarRequest]) (*connect.Response[v1.SetVarResponse], error)
	RemoveVar(context.Context, *connect.Request[v1.RemoveVarRequest]) (*connect.Response[v1.RemoveVarResponse], error)
	GetVars(context.Context, *connect.Request[v1.GetVarsRequest]) (*connect.Response[v1.GetVarsResponse], error)
	RevealVar(context.Context, *connect.Request[v1.RevealVarRequest]) (*connect.Response[v1.RevealVarResponse], error)
}

// NewEnvsServiceClient constructs a client for the autokitteh.envs.v1.EnvsService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewEnvsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) EnvsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &envsServiceClient{
		list: connect.NewClient[v1.ListRequest, v1.ListResponse](
			httpClient,
			baseURL+EnvsServiceListProcedure,
			opts...,
		),
		create: connect.NewClient[v1.CreateRequest, v1.CreateResponse](
			httpClient,
			baseURL+EnvsServiceCreateProcedure,
			opts...,
		),
		get: connect.NewClient[v1.GetRequest, v1.GetResponse](
			httpClient,
			baseURL+EnvsServiceGetProcedure,
			opts...,
		),
		remove: connect.NewClient[v1.RemoveRequest, v1.RemoveResponse](
			httpClient,
			baseURL+EnvsServiceRemoveProcedure,
			opts...,
		),
		update: connect.NewClient[v1.UpdateRequest, v1.UpdateResponse](
			httpClient,
			baseURL+EnvsServiceUpdateProcedure,
			opts...,
		),
		setVar: connect.NewClient[v1.SetVarRequest, v1.SetVarResponse](
			httpClient,
			baseURL+EnvsServiceSetVarProcedure,
			opts...,
		),
		removeVar: connect.NewClient[v1.RemoveVarRequest, v1.RemoveVarResponse](
			httpClient,
			baseURL+EnvsServiceRemoveVarProcedure,
			opts...,
		),
		getVars: connect.NewClient[v1.GetVarsRequest, v1.GetVarsResponse](
			httpClient,
			baseURL+EnvsServiceGetVarsProcedure,
			opts...,
		),
		revealVar: connect.NewClient[v1.RevealVarRequest, v1.RevealVarResponse](
			httpClient,
			baseURL+EnvsServiceRevealVarProcedure,
			opts...,
		),
	}
}

// envsServiceClient implements EnvsServiceClient.
type envsServiceClient struct {
	list      *connect.Client[v1.ListRequest, v1.ListResponse]
	create    *connect.Client[v1.CreateRequest, v1.CreateResponse]
	get       *connect.Client[v1.GetRequest, v1.GetResponse]
	remove    *connect.Client[v1.RemoveRequest, v1.RemoveResponse]
	update    *connect.Client[v1.UpdateRequest, v1.UpdateResponse]
	setVar    *connect.Client[v1.SetVarRequest, v1.SetVarResponse]
	removeVar *connect.Client[v1.RemoveVarRequest, v1.RemoveVarResponse]
	getVars   *connect.Client[v1.GetVarsRequest, v1.GetVarsResponse]
	revealVar *connect.Client[v1.RevealVarRequest, v1.RevealVarResponse]
}

// List calls autokitteh.envs.v1.EnvsService.List.
func (c *envsServiceClient) List(ctx context.Context, req *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return c.list.CallUnary(ctx, req)
}

// Create calls autokitteh.envs.v1.EnvsService.Create.
func (c *envsServiceClient) Create(ctx context.Context, req *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error) {
	return c.create.CallUnary(ctx, req)
}

// Get calls autokitteh.envs.v1.EnvsService.Get.
func (c *envsServiceClient) Get(ctx context.Context, req *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// Remove calls autokitteh.envs.v1.EnvsService.Remove.
func (c *envsServiceClient) Remove(ctx context.Context, req *connect.Request[v1.RemoveRequest]) (*connect.Response[v1.RemoveResponse], error) {
	return c.remove.CallUnary(ctx, req)
}

// Update calls autokitteh.envs.v1.EnvsService.Update.
func (c *envsServiceClient) Update(ctx context.Context, req *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error) {
	return c.update.CallUnary(ctx, req)
}

// SetVar calls autokitteh.envs.v1.EnvsService.SetVar.
func (c *envsServiceClient) SetVar(ctx context.Context, req *connect.Request[v1.SetVarRequest]) (*connect.Response[v1.SetVarResponse], error) {
	return c.setVar.CallUnary(ctx, req)
}

// RemoveVar calls autokitteh.envs.v1.EnvsService.RemoveVar.
func (c *envsServiceClient) RemoveVar(ctx context.Context, req *connect.Request[v1.RemoveVarRequest]) (*connect.Response[v1.RemoveVarResponse], error) {
	return c.removeVar.CallUnary(ctx, req)
}

// GetVars calls autokitteh.envs.v1.EnvsService.GetVars.
func (c *envsServiceClient) GetVars(ctx context.Context, req *connect.Request[v1.GetVarsRequest]) (*connect.Response[v1.GetVarsResponse], error) {
	return c.getVars.CallUnary(ctx, req)
}

// RevealVar calls autokitteh.envs.v1.EnvsService.RevealVar.
func (c *envsServiceClient) RevealVar(ctx context.Context, req *connect.Request[v1.RevealVarRequest]) (*connect.Response[v1.RevealVarResponse], error) {
	return c.revealVar.CallUnary(ctx, req)
}

// EnvsServiceHandler is an implementation of the autokitteh.envs.v1.EnvsService service.
type EnvsServiceHandler interface {
	List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error)
	Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error)
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	Remove(context.Context, *connect.Request[v1.RemoveRequest]) (*connect.Response[v1.RemoveResponse], error)
	Update(context.Context, *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error)
	SetVar(context.Context, *connect.Request[v1.SetVarRequest]) (*connect.Response[v1.SetVarResponse], error)
	RemoveVar(context.Context, *connect.Request[v1.RemoveVarRequest]) (*connect.Response[v1.RemoveVarResponse], error)
	GetVars(context.Context, *connect.Request[v1.GetVarsRequest]) (*connect.Response[v1.GetVarsResponse], error)
	RevealVar(context.Context, *connect.Request[v1.RevealVarRequest]) (*connect.Response[v1.RevealVarResponse], error)
}

// NewEnvsServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewEnvsServiceHandler(svc EnvsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	envsServiceListHandler := connect.NewUnaryHandler(
		EnvsServiceListProcedure,
		svc.List,
		opts...,
	)
	envsServiceCreateHandler := connect.NewUnaryHandler(
		EnvsServiceCreateProcedure,
		svc.Create,
		opts...,
	)
	envsServiceGetHandler := connect.NewUnaryHandler(
		EnvsServiceGetProcedure,
		svc.Get,
		opts...,
	)
	envsServiceRemoveHandler := connect.NewUnaryHandler(
		EnvsServiceRemoveProcedure,
		svc.Remove,
		opts...,
	)
	envsServiceUpdateHandler := connect.NewUnaryHandler(
		EnvsServiceUpdateProcedure,
		svc.Update,
		opts...,
	)
	envsServiceSetVarHandler := connect.NewUnaryHandler(
		EnvsServiceSetVarProcedure,
		svc.SetVar,
		opts...,
	)
	envsServiceRemoveVarHandler := connect.NewUnaryHandler(
		EnvsServiceRemoveVarProcedure,
		svc.RemoveVar,
		opts...,
	)
	envsServiceGetVarsHandler := connect.NewUnaryHandler(
		EnvsServiceGetVarsProcedure,
		svc.GetVars,
		opts...,
	)
	envsServiceRevealVarHandler := connect.NewUnaryHandler(
		EnvsServiceRevealVarProcedure,
		svc.RevealVar,
		opts...,
	)
	return "/autokitteh.envs.v1.EnvsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case EnvsServiceListProcedure:
			envsServiceListHandler.ServeHTTP(w, r)
		case EnvsServiceCreateProcedure:
			envsServiceCreateHandler.ServeHTTP(w, r)
		case EnvsServiceGetProcedure:
			envsServiceGetHandler.ServeHTTP(w, r)
		case EnvsServiceRemoveProcedure:
			envsServiceRemoveHandler.ServeHTTP(w, r)
		case EnvsServiceUpdateProcedure:
			envsServiceUpdateHandler.ServeHTTP(w, r)
		case EnvsServiceSetVarProcedure:
			envsServiceSetVarHandler.ServeHTTP(w, r)
		case EnvsServiceRemoveVarProcedure:
			envsServiceRemoveVarHandler.ServeHTTP(w, r)
		case EnvsServiceGetVarsProcedure:
			envsServiceGetVarsHandler.ServeHTTP(w, r)
		case EnvsServiceRevealVarProcedure:
			envsServiceRevealVarHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedEnvsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedEnvsServiceHandler struct{}

func (UnimplementedEnvsServiceHandler) List(context.Context, *connect.Request[v1.ListRequest]) (*connect.Response[v1.ListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.envs.v1.EnvsService.List is not implemented"))
}

func (UnimplementedEnvsServiceHandler) Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.envs.v1.EnvsService.Create is not implemented"))
}

func (UnimplementedEnvsServiceHandler) Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.envs.v1.EnvsService.Get is not implemented"))
}

func (UnimplementedEnvsServiceHandler) Remove(context.Context, *connect.Request[v1.RemoveRequest]) (*connect.Response[v1.RemoveResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.envs.v1.EnvsService.Remove is not implemented"))
}

func (UnimplementedEnvsServiceHandler) Update(context.Context, *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.envs.v1.EnvsService.Update is not implemented"))
}

func (UnimplementedEnvsServiceHandler) SetVar(context.Context, *connect.Request[v1.SetVarRequest]) (*connect.Response[v1.SetVarResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.envs.v1.EnvsService.SetVar is not implemented"))
}

func (UnimplementedEnvsServiceHandler) RemoveVar(context.Context, *connect.Request[v1.RemoveVarRequest]) (*connect.Response[v1.RemoveVarResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.envs.v1.EnvsService.RemoveVar is not implemented"))
}

func (UnimplementedEnvsServiceHandler) GetVars(context.Context, *connect.Request[v1.GetVarsRequest]) (*connect.Response[v1.GetVarsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.envs.v1.EnvsService.GetVars is not implemented"))
}

func (UnimplementedEnvsServiceHandler) RevealVar(context.Context, *connect.Request[v1.RevealVarRequest]) (*connect.Response[v1.RevealVarResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.envs.v1.EnvsService.RevealVar is not implemented"))
}
