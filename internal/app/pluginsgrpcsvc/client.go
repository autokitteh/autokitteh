package pluginsgrpcsvc

import (
	"context"

	"google.golang.org/grpc"

	pb "github.com/autokitteh/autokitteh/gen/proto/stubs/go/pluginsprovidersvc"
)

type LocalClient struct{ Server pb.PluginsProviderServer }

var _ pb.PluginsProviderClient = &LocalClient{}

func (c *LocalClient) List(ctx context.Context, req *pb.ListRequest, _ ...grpc.CallOption) (*pb.ListResponse, error) {
	return c.Server.List(ctx, req)
}

func (c *LocalClient) GetValues(ctx context.Context, req *pb.GetValuesRequest, _ ...grpc.CallOption) (*pb.GetValuesResponse, error) {
	return c.Server.GetValues(ctx, req)
}

func (c *LocalClient) Describe(ctx context.Context, req *pb.DescribeRequest, _ ...grpc.CallOption) (*pb.DescribeResponse, error) {
	return c.Server.Describe(ctx, req)
}

func (c *LocalClient) CallValue(ctx context.Context, req *pb.CallValueRequest, _ ...grpc.CallOption) (*pb.CallValueResponse, error) {
	return c.Server.CallValue(ctx, req)
}
