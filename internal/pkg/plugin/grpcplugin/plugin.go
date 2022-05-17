package grpcplugin

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/autokitteh/autokitteh/api/gen/stubs/go/pluginsprovidersvc"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiplugin"
	"github.com/autokitteh/autokitteh/internal/pkg/plugin"
	L "github.com/autokitteh/autokitteh/pkg/l"
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
