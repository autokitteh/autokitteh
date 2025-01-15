// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: autokitteh/orgs/v1/svc.proto

package orgsv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1"
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
	// OrgsServiceName is the fully-qualified name of the OrgsService service.
	OrgsServiceName = "autokitteh.orgs.v1.OrgsService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// OrgsServiceCreateProcedure is the fully-qualified name of the OrgsService's Create RPC.
	OrgsServiceCreateProcedure = "/autokitteh.orgs.v1.OrgsService/Create"
	// OrgsServiceGetProcedure is the fully-qualified name of the OrgsService's Get RPC.
	OrgsServiceGetProcedure = "/autokitteh.orgs.v1.OrgsService/Get"
	// OrgsServiceBatchGetProcedure is the fully-qualified name of the OrgsService's BatchGet RPC.
	OrgsServiceBatchGetProcedure = "/autokitteh.orgs.v1.OrgsService/BatchGet"
	// OrgsServiceUpdateProcedure is the fully-qualified name of the OrgsService's Update RPC.
	OrgsServiceUpdateProcedure = "/autokitteh.orgs.v1.OrgsService/Update"
	// OrgsServiceDeleteProcedure is the fully-qualified name of the OrgsService's Delete RPC.
	OrgsServiceDeleteProcedure = "/autokitteh.orgs.v1.OrgsService/Delete"
	// OrgsServiceAddMemberProcedure is the fully-qualified name of the OrgsService's AddMember RPC.
	OrgsServiceAddMemberProcedure = "/autokitteh.orgs.v1.OrgsService/AddMember"
	// OrgsServiceUpdateMemberProcedure is the fully-qualified name of the OrgsService's UpdateMember
	// RPC.
	OrgsServiceUpdateMemberProcedure = "/autokitteh.orgs.v1.OrgsService/UpdateMember"
	// OrgsServiceRemoveMemberProcedure is the fully-qualified name of the OrgsService's RemoveMember
	// RPC.
	OrgsServiceRemoveMemberProcedure = "/autokitteh.orgs.v1.OrgsService/RemoveMember"
	// OrgsServiceListMembersProcedure is the fully-qualified name of the OrgsService's ListMembers RPC.
	OrgsServiceListMembersProcedure = "/autokitteh.orgs.v1.OrgsService/ListMembers"
	// OrgsServiceGetMemberProcedure is the fully-qualified name of the OrgsService's GetMember RPC.
	OrgsServiceGetMemberProcedure = "/autokitteh.orgs.v1.OrgsService/GetMember"
	// OrgsServiceGetOrgsForUserProcedure is the fully-qualified name of the OrgsService's
	// GetOrgsForUser RPC.
	OrgsServiceGetOrgsForUserProcedure = "/autokitteh.orgs.v1.OrgsService/GetOrgsForUser"
)

// OrgsServiceClient is a client for the autokitteh.orgs.v1.OrgsService service.
type OrgsServiceClient interface {
	Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error)
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	// BatchGet returns a list of orgs for the given org_ids, if the org does not exist, it will not be returned.
	BatchGet(context.Context, *connect.Request[v1.BatchGetRequest]) (*connect.Response[v1.BatchGetResponse], error)
	Update(context.Context, *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error)
	Delete(context.Context, *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error)
	AddMember(context.Context, *connect.Request[v1.AddMemberRequest]) (*connect.Response[v1.AddMemberResponse], error)
	UpdateMember(context.Context, *connect.Request[v1.UpdateMemberRequest]) (*connect.Response[v1.UpdateMemberResponse], error)
	RemoveMember(context.Context, *connect.Request[v1.RemoveMemberRequest]) (*connect.Response[v1.RemoveMemberResponse], error)
	ListMembers(context.Context, *connect.Request[v1.ListMembersRequest]) (*connect.Response[v1.ListMembersResponse], error)
	GetMember(context.Context, *connect.Request[v1.GetMemberRequest]) (*connect.Response[v1.GetMemberResponse], error)
	GetOrgsForUser(context.Context, *connect.Request[v1.GetOrgsForUserRequest]) (*connect.Response[v1.GetOrgsForUserResponse], error)
}

// NewOrgsServiceClient constructs a client for the autokitteh.orgs.v1.OrgsService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewOrgsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) OrgsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &orgsServiceClient{
		create: connect.NewClient[v1.CreateRequest, v1.CreateResponse](
			httpClient,
			baseURL+OrgsServiceCreateProcedure,
			opts...,
		),
		get: connect.NewClient[v1.GetRequest, v1.GetResponse](
			httpClient,
			baseURL+OrgsServiceGetProcedure,
			opts...,
		),
		batchGet: connect.NewClient[v1.BatchGetRequest, v1.BatchGetResponse](
			httpClient,
			baseURL+OrgsServiceBatchGetProcedure,
			opts...,
		),
		update: connect.NewClient[v1.UpdateRequest, v1.UpdateResponse](
			httpClient,
			baseURL+OrgsServiceUpdateProcedure,
			opts...,
		),
		delete: connect.NewClient[v1.DeleteRequest, v1.DeleteResponse](
			httpClient,
			baseURL+OrgsServiceDeleteProcedure,
			opts...,
		),
		addMember: connect.NewClient[v1.AddMemberRequest, v1.AddMemberResponse](
			httpClient,
			baseURL+OrgsServiceAddMemberProcedure,
			opts...,
		),
		updateMember: connect.NewClient[v1.UpdateMemberRequest, v1.UpdateMemberResponse](
			httpClient,
			baseURL+OrgsServiceUpdateMemberProcedure,
			opts...,
		),
		removeMember: connect.NewClient[v1.RemoveMemberRequest, v1.RemoveMemberResponse](
			httpClient,
			baseURL+OrgsServiceRemoveMemberProcedure,
			opts...,
		),
		listMembers: connect.NewClient[v1.ListMembersRequest, v1.ListMembersResponse](
			httpClient,
			baseURL+OrgsServiceListMembersProcedure,
			opts...,
		),
		getMember: connect.NewClient[v1.GetMemberRequest, v1.GetMemberResponse](
			httpClient,
			baseURL+OrgsServiceGetMemberProcedure,
			opts...,
		),
		getOrgsForUser: connect.NewClient[v1.GetOrgsForUserRequest, v1.GetOrgsForUserResponse](
			httpClient,
			baseURL+OrgsServiceGetOrgsForUserProcedure,
			opts...,
		),
	}
}

// orgsServiceClient implements OrgsServiceClient.
type orgsServiceClient struct {
	create         *connect.Client[v1.CreateRequest, v1.CreateResponse]
	get            *connect.Client[v1.GetRequest, v1.GetResponse]
	batchGet       *connect.Client[v1.BatchGetRequest, v1.BatchGetResponse]
	update         *connect.Client[v1.UpdateRequest, v1.UpdateResponse]
	delete         *connect.Client[v1.DeleteRequest, v1.DeleteResponse]
	addMember      *connect.Client[v1.AddMemberRequest, v1.AddMemberResponse]
	updateMember   *connect.Client[v1.UpdateMemberRequest, v1.UpdateMemberResponse]
	removeMember   *connect.Client[v1.RemoveMemberRequest, v1.RemoveMemberResponse]
	listMembers    *connect.Client[v1.ListMembersRequest, v1.ListMembersResponse]
	getMember      *connect.Client[v1.GetMemberRequest, v1.GetMemberResponse]
	getOrgsForUser *connect.Client[v1.GetOrgsForUserRequest, v1.GetOrgsForUserResponse]
}

// Create calls autokitteh.orgs.v1.OrgsService.Create.
func (c *orgsServiceClient) Create(ctx context.Context, req *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error) {
	return c.create.CallUnary(ctx, req)
}

// Get calls autokitteh.orgs.v1.OrgsService.Get.
func (c *orgsServiceClient) Get(ctx context.Context, req *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return c.get.CallUnary(ctx, req)
}

// BatchGet calls autokitteh.orgs.v1.OrgsService.BatchGet.
func (c *orgsServiceClient) BatchGet(ctx context.Context, req *connect.Request[v1.BatchGetRequest]) (*connect.Response[v1.BatchGetResponse], error) {
	return c.batchGet.CallUnary(ctx, req)
}

// Update calls autokitteh.orgs.v1.OrgsService.Update.
func (c *orgsServiceClient) Update(ctx context.Context, req *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error) {
	return c.update.CallUnary(ctx, req)
}

// Delete calls autokitteh.orgs.v1.OrgsService.Delete.
func (c *orgsServiceClient) Delete(ctx context.Context, req *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error) {
	return c.delete.CallUnary(ctx, req)
}

// AddMember calls autokitteh.orgs.v1.OrgsService.AddMember.
func (c *orgsServiceClient) AddMember(ctx context.Context, req *connect.Request[v1.AddMemberRequest]) (*connect.Response[v1.AddMemberResponse], error) {
	return c.addMember.CallUnary(ctx, req)
}

// UpdateMember calls autokitteh.orgs.v1.OrgsService.UpdateMember.
func (c *orgsServiceClient) UpdateMember(ctx context.Context, req *connect.Request[v1.UpdateMemberRequest]) (*connect.Response[v1.UpdateMemberResponse], error) {
	return c.updateMember.CallUnary(ctx, req)
}

// RemoveMember calls autokitteh.orgs.v1.OrgsService.RemoveMember.
func (c *orgsServiceClient) RemoveMember(ctx context.Context, req *connect.Request[v1.RemoveMemberRequest]) (*connect.Response[v1.RemoveMemberResponse], error) {
	return c.removeMember.CallUnary(ctx, req)
}

// ListMembers calls autokitteh.orgs.v1.OrgsService.ListMembers.
func (c *orgsServiceClient) ListMembers(ctx context.Context, req *connect.Request[v1.ListMembersRequest]) (*connect.Response[v1.ListMembersResponse], error) {
	return c.listMembers.CallUnary(ctx, req)
}

// GetMember calls autokitteh.orgs.v1.OrgsService.GetMember.
func (c *orgsServiceClient) GetMember(ctx context.Context, req *connect.Request[v1.GetMemberRequest]) (*connect.Response[v1.GetMemberResponse], error) {
	return c.getMember.CallUnary(ctx, req)
}

// GetOrgsForUser calls autokitteh.orgs.v1.OrgsService.GetOrgsForUser.
func (c *orgsServiceClient) GetOrgsForUser(ctx context.Context, req *connect.Request[v1.GetOrgsForUserRequest]) (*connect.Response[v1.GetOrgsForUserResponse], error) {
	return c.getOrgsForUser.CallUnary(ctx, req)
}

// OrgsServiceHandler is an implementation of the autokitteh.orgs.v1.OrgsService service.
type OrgsServiceHandler interface {
	Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error)
	Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error)
	// BatchGet returns a list of orgs for the given org_ids, if the org does not exist, it will not be returned.
	BatchGet(context.Context, *connect.Request[v1.BatchGetRequest]) (*connect.Response[v1.BatchGetResponse], error)
	Update(context.Context, *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error)
	Delete(context.Context, *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error)
	AddMember(context.Context, *connect.Request[v1.AddMemberRequest]) (*connect.Response[v1.AddMemberResponse], error)
	UpdateMember(context.Context, *connect.Request[v1.UpdateMemberRequest]) (*connect.Response[v1.UpdateMemberResponse], error)
	RemoveMember(context.Context, *connect.Request[v1.RemoveMemberRequest]) (*connect.Response[v1.RemoveMemberResponse], error)
	ListMembers(context.Context, *connect.Request[v1.ListMembersRequest]) (*connect.Response[v1.ListMembersResponse], error)
	GetMember(context.Context, *connect.Request[v1.GetMemberRequest]) (*connect.Response[v1.GetMemberResponse], error)
	GetOrgsForUser(context.Context, *connect.Request[v1.GetOrgsForUserRequest]) (*connect.Response[v1.GetOrgsForUserResponse], error)
}

// NewOrgsServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewOrgsServiceHandler(svc OrgsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	orgsServiceCreateHandler := connect.NewUnaryHandler(
		OrgsServiceCreateProcedure,
		svc.Create,
		opts...,
	)
	orgsServiceGetHandler := connect.NewUnaryHandler(
		OrgsServiceGetProcedure,
		svc.Get,
		opts...,
	)
	orgsServiceBatchGetHandler := connect.NewUnaryHandler(
		OrgsServiceBatchGetProcedure,
		svc.BatchGet,
		opts...,
	)
	orgsServiceUpdateHandler := connect.NewUnaryHandler(
		OrgsServiceUpdateProcedure,
		svc.Update,
		opts...,
	)
	orgsServiceDeleteHandler := connect.NewUnaryHandler(
		OrgsServiceDeleteProcedure,
		svc.Delete,
		opts...,
	)
	orgsServiceAddMemberHandler := connect.NewUnaryHandler(
		OrgsServiceAddMemberProcedure,
		svc.AddMember,
		opts...,
	)
	orgsServiceUpdateMemberHandler := connect.NewUnaryHandler(
		OrgsServiceUpdateMemberProcedure,
		svc.UpdateMember,
		opts...,
	)
	orgsServiceRemoveMemberHandler := connect.NewUnaryHandler(
		OrgsServiceRemoveMemberProcedure,
		svc.RemoveMember,
		opts...,
	)
	orgsServiceListMembersHandler := connect.NewUnaryHandler(
		OrgsServiceListMembersProcedure,
		svc.ListMembers,
		opts...,
	)
	orgsServiceGetMemberHandler := connect.NewUnaryHandler(
		OrgsServiceGetMemberProcedure,
		svc.GetMember,
		opts...,
	)
	orgsServiceGetOrgsForUserHandler := connect.NewUnaryHandler(
		OrgsServiceGetOrgsForUserProcedure,
		svc.GetOrgsForUser,
		opts...,
	)
	return "/autokitteh.orgs.v1.OrgsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case OrgsServiceCreateProcedure:
			orgsServiceCreateHandler.ServeHTTP(w, r)
		case OrgsServiceGetProcedure:
			orgsServiceGetHandler.ServeHTTP(w, r)
		case OrgsServiceBatchGetProcedure:
			orgsServiceBatchGetHandler.ServeHTTP(w, r)
		case OrgsServiceUpdateProcedure:
			orgsServiceUpdateHandler.ServeHTTP(w, r)
		case OrgsServiceDeleteProcedure:
			orgsServiceDeleteHandler.ServeHTTP(w, r)
		case OrgsServiceAddMemberProcedure:
			orgsServiceAddMemberHandler.ServeHTTP(w, r)
		case OrgsServiceUpdateMemberProcedure:
			orgsServiceUpdateMemberHandler.ServeHTTP(w, r)
		case OrgsServiceRemoveMemberProcedure:
			orgsServiceRemoveMemberHandler.ServeHTTP(w, r)
		case OrgsServiceListMembersProcedure:
			orgsServiceListMembersHandler.ServeHTTP(w, r)
		case OrgsServiceGetMemberProcedure:
			orgsServiceGetMemberHandler.ServeHTTP(w, r)
		case OrgsServiceGetOrgsForUserProcedure:
			orgsServiceGetOrgsForUserHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedOrgsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedOrgsServiceHandler struct{}

func (UnimplementedOrgsServiceHandler) Create(context.Context, *connect.Request[v1.CreateRequest]) (*connect.Response[v1.CreateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.Create is not implemented"))
}

func (UnimplementedOrgsServiceHandler) Get(context.Context, *connect.Request[v1.GetRequest]) (*connect.Response[v1.GetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.Get is not implemented"))
}

func (UnimplementedOrgsServiceHandler) BatchGet(context.Context, *connect.Request[v1.BatchGetRequest]) (*connect.Response[v1.BatchGetResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.BatchGet is not implemented"))
}

func (UnimplementedOrgsServiceHandler) Update(context.Context, *connect.Request[v1.UpdateRequest]) (*connect.Response[v1.UpdateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.Update is not implemented"))
}

func (UnimplementedOrgsServiceHandler) Delete(context.Context, *connect.Request[v1.DeleteRequest]) (*connect.Response[v1.DeleteResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.Delete is not implemented"))
}

func (UnimplementedOrgsServiceHandler) AddMember(context.Context, *connect.Request[v1.AddMemberRequest]) (*connect.Response[v1.AddMemberResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.AddMember is not implemented"))
}

func (UnimplementedOrgsServiceHandler) UpdateMember(context.Context, *connect.Request[v1.UpdateMemberRequest]) (*connect.Response[v1.UpdateMemberResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.UpdateMember is not implemented"))
}

func (UnimplementedOrgsServiceHandler) RemoveMember(context.Context, *connect.Request[v1.RemoveMemberRequest]) (*connect.Response[v1.RemoveMemberResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.RemoveMember is not implemented"))
}

func (UnimplementedOrgsServiceHandler) ListMembers(context.Context, *connect.Request[v1.ListMembersRequest]) (*connect.Response[v1.ListMembersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.ListMembers is not implemented"))
}

func (UnimplementedOrgsServiceHandler) GetMember(context.Context, *connect.Request[v1.GetMemberRequest]) (*connect.Response[v1.GetMemberResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.GetMember is not implemented"))
}

func (UnimplementedOrgsServiceHandler) GetOrgsForUser(context.Context, *connect.Request[v1.GetOrgsForUserRequest]) (*connect.Response[v1.GetOrgsForUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("autokitteh.orgs.v1.OrgsService.GetOrgsForUser is not implemented"))
}
