package eventsstoregrpcsvc

import (
	"context"

	"google.golang.org/grpc"

	pbeventsvc "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/eventsvc"
)

type LocalClient struct {
	Server pbeventsvc.EventsServer
}

var _ pbeventsvc.EventsClient = &LocalClient{}

func (c *LocalClient) IngestEvent(ctx context.Context, req *pbeventsvc.IngestEventRequest, _ ...grpc.CallOption) (*pbeventsvc.IngestEventResponse, error) {
	return c.Server.IngestEvent(ctx, req)
}

func (c *LocalClient) GetEvent(ctx context.Context, req *pbeventsvc.GetEventRequest, _ ...grpc.CallOption) (*pbeventsvc.GetEventResponse, error) {
	return c.Server.GetEvent(ctx, req)
}

func (c *LocalClient) ListEvents(ctx context.Context, req *pbeventsvc.ListEventsRequest, _ ...grpc.CallOption) (*pbeventsvc.ListEventsResponse, error) {
	return c.Server.ListEvents(ctx, req)
}

func (c *LocalClient) GetEventState(ctx context.Context, in *pbeventsvc.GetEventStateRequest, _ ...grpc.CallOption) (*pbeventsvc.GetEventStateResponse, error) {
	return c.Server.GetEventState(ctx, in)
}

func (c *LocalClient) UpdateEventState(ctx context.Context, in *pbeventsvc.UpdateEventStateRequest, _ ...grpc.CallOption) (*pbeventsvc.UpdateEventStateResponse, error) {
	return c.Server.UpdateEventState(ctx, in)
}

func (c *LocalClient) GetEventStateForProject(ctx context.Context, req *pbeventsvc.GetEventStateForProjectRequest, _ ...grpc.CallOption) (*pbeventsvc.GetEventStateForProjectResponse, error) {
	return c.Server.GetEventStateForProject(ctx, req)
}

func (c *LocalClient) UpdateEventStateForProject(ctx context.Context, req *pbeventsvc.UpdateEventStateForProjectRequest, _ ...grpc.CallOption) (*pbeventsvc.UpdateEventStateForProjectResponse, error) {
	return c.Server.UpdateEventStateForProject(ctx, req)
}

func (c *LocalClient) GetProjectWaitingEvents(ctx context.Context, req *pbeventsvc.GetProjectWaitingEventsRequest, _ ...grpc.CallOption) (*pbeventsvc.GetProjectWaitingEventsResponse, error) {
	return c.Server.GetProjectWaitingEvents(ctx, req)
}
