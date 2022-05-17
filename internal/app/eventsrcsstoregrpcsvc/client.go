package eventsrcsstoregrpcsvc

import (
	"context"

	"google.golang.org/grpc"

	pbeventsrcsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/eventsrcsvc"
)

type LocalClient struct {
	Server pbeventsrcsvc.EventSourcesServer
}

var _ pbeventsrcsvc.EventSourcesClient = &LocalClient{}

func (c *LocalClient) AddEventSource(ctx context.Context, in *pbeventsrcsvc.AddEventSourceRequest, _ ...grpc.CallOption) (*pbeventsrcsvc.AddEventSourceResponse, error) {
	return c.Server.AddEventSource(ctx, in)
}

func (c *LocalClient) UpdateEventSource(ctx context.Context, in *pbeventsrcsvc.UpdateEventSourceRequest, _ ...grpc.CallOption) (*pbeventsrcsvc.UpdateEventSourceResponse, error) {
	return c.Server.UpdateEventSource(ctx, in)
}

func (c *LocalClient) GetEventSource(ctx context.Context, in *pbeventsrcsvc.GetEventSourceRequest, _ ...grpc.CallOption) (*pbeventsrcsvc.GetEventSourceResponse, error) {
	return c.Server.GetEventSource(ctx, in)
}

func (c *LocalClient) AddEventSourceProjectBinding(ctx context.Context, in *pbeventsrcsvc.AddEventSourceProjectBindingRequest, _ ...grpc.CallOption) (*pbeventsrcsvc.AddEventSourceProjectBindingResponse, error) {
	return c.Server.AddEventSourceProjectBinding(ctx, in)
}

func (c *LocalClient) UpdateEventSourceProjectBinding(ctx context.Context, in *pbeventsrcsvc.UpdateEventSourceProjectBindingRequest, _ ...grpc.CallOption) (*pbeventsrcsvc.UpdateEventSourceProjectBindingResponse, error) {
	return c.Server.UpdateEventSourceProjectBinding(ctx, in)
}

func (c *LocalClient) GetEventSourceProjectBindings(ctx context.Context, in *pbeventsrcsvc.GetEventSourceProjectBindingsRequest, _ ...grpc.CallOption) (*pbeventsrcsvc.GetEventSourceProjectBindingsResponse, error) {
	return c.Server.GetEventSourceProjectBindings(ctx, in)
}

func (c *LocalClient) ListEventSources(ctx context.Context, in *pbeventsrcsvc.ListEventSourcesRequest, _ ...grpc.CallOption) (*pbeventsrcsvc.ListEventSourcesResponse, error) {
	return c.Server.ListEventSources(ctx, in)
}
