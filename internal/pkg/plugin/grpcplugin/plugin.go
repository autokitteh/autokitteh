package grpcplugin

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/pluginsprovidersvc"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiplugin"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/plugin"
	L "gitlab.com/softkitteh/autokitteh/pkg/l"
)

type GRPCPlugin struct {
	L      L.Nullable
	ID     apiplugin.PluginID
	Client pb.PluginsProviderClient
}

var _ plugin.Plugin = &GRPCPlugin{}

func NewFromHostPort(l L.L, id apiplugin.PluginID, hostPort string) (plugin.Plugin, error) {
	conn, err := grpc.Dial(hostPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}

	return NewFromGRPCConn(l, id, conn)
}

func NewFromGRPCConn(l L.L, id apiplugin.PluginID, conn *grpc.ClientConn) (plugin.Plugin, error) {
	return &GRPCPlugin{
		ID:     id,
		L:      L.N(L.N(l).Named("grpcplugin").With("id", id)),
		Client: pb.NewPluginsProviderClient(conn),
	}, nil
}
