package projectsstoregrpcsvc

import (
	"context"

	"google.golang.org/grpc"

	pbprojectsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/projectsvc"
)

type LocalClient struct {
	Server pbprojectsvc.ProjectsServer
}

var _ pbprojectsvc.ProjectsClient = &LocalClient{}

func (c *LocalClient) CreateProject(ctx context.Context, in *pbprojectsvc.CreateProjectRequest, _ ...grpc.CallOption) (*pbprojectsvc.CreateProjectResponse, error) {
	return c.Server.CreateProject(ctx, in)
}

func (c *LocalClient) UpdateProject(ctx context.Context, in *pbprojectsvc.UpdateProjectRequest, _ ...grpc.CallOption) (*pbprojectsvc.UpdateProjectResponse, error) {
	return c.Server.UpdateProject(ctx, in)
}

func (c *LocalClient) GetProject(ctx context.Context, in *pbprojectsvc.GetProjectRequest, _ ...grpc.CallOption) (*pbprojectsvc.GetProjectResponse, error) {
	return c.Server.GetProject(ctx, in)
}

func (c *LocalClient) GetProjects(ctx context.Context, in *pbprojectsvc.GetProjectsRequest, _ ...grpc.CallOption) (*pbprojectsvc.GetProjectsResponse, error) {
	return c.Server.GetProjects(ctx, in)
}
