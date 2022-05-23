package pluginsreggrpcsvc

import (
	"context"

	"google.golang.org/grpc"

	pb "go.autokitteh.dev/idl/go/pluginsregsvc"
)

type LocalClient struct{ Server pb.PluginsRegistryServer }

var _ pb.PluginsRegistryClient = &LocalClient{}

func (c *LocalClient) List(ctx context.Context, req *pb.ListRequest, _ ...grpc.CallOption) (*pb.ListResponse, error) {
	return c.Server.List(ctx, req)
}

func (c *LocalClient) Get(ctx context.Context, req *pb.GetRequest, _ ...grpc.CallOption) (*pb.GetResponse, error) {
	return c.Server.Get(ctx, req)
}

func (c *LocalClient) Register(ctx context.Context, req *pb.RegisterRequest, _ ...grpc.CallOption) (*pb.RegisterResponse, error) {
	return c.Server.Register(ctx, req)
}
