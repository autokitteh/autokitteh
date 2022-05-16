package langgrpcsvc

import (
	"context"

	"google.golang.org/grpc"

	pblangsvc "github.com/autokitteh/autokitteh/gen/proto/stubs/go/langsvc"
)

type LocalClient struct {
	Server pblangsvc.LangServer
}

var _ pblangsvc.LangClient = &LocalClient{}

func (c *LocalClient) ListLangs(ctx context.Context, in *pblangsvc.ListLangsRequest, _ ...grpc.CallOption) (*pblangsvc.ListLangsResponse, error) {
	return c.Server.ListLangs(ctx, in)
}

func (c *LocalClient) IsCompilerVersionSupported(ctx context.Context, in *pblangsvc.IsCompilerVersionSupportedRequest, _ ...grpc.CallOption) (*pblangsvc.IsCompilerVersionSupportedResponse, error) {
	return c.Server.IsCompilerVersionSupported(ctx, in)
}

func (c *LocalClient) GetModuleDependencies(ctx context.Context, in *pblangsvc.GetModuleDependenciesRequest, _ ...grpc.CallOption) (*pblangsvc.GetModuleDependenciesResponse, error) {
	return c.Server.GetModuleDependencies(ctx, in)
}

func (c *LocalClient) CompileModule(ctx context.Context, in *pblangsvc.CompileModuleRequest, _ ...grpc.CallOption) (*pblangsvc.CompileModuleResponse, error) {
	return c.Server.CompileModule(ctx, in)
}
