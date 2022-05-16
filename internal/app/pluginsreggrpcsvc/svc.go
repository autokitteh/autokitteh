package pluginsreggrpcsvc

import (
	"context"
	"errors"

	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbsvc "github.com/autokitteh/autokitteh/gen/proto/stubs/go/pluginsregsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/pluginsreg"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiplugin"

	L "github.com/autokitteh/autokitteh/pkg/l"
)

type Svc struct {
	pbsvc.UnimplementedPluginsRegistryServer

	Registry *pluginsreg.Registry

	L L.Nullable
}

var _ pbsvc.PluginsRegistryServer = &Svc{}

func (s *Svc) RegisterServer(srv *grpc.Server) { pbsvc.RegisterPluginsRegistryServer(srv, s) }

func (s *Svc) List(ctx context.Context, req *pbsvc.ListRequest) (*pbsvc.ListResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	a_ := apiaccount.AccountName(req.AccountName)
	a := &a_
	if a.Empty() {
		a = nil
	}

	ids, err := s.Registry.List(ctx, a)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "%v", err)
	}

	pbids := lo.Map(ids, func(id apiplugin.PluginID, _ int) string { return id.String() })

	return &pbsvc.ListResponse{Ids: pbids}, nil
}

func (s *Svc) Get(ctx context.Context, req *pbsvc.GetRequest) (*pbsvc.GetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	p, err := s.Registry.Get(ctx, apiplugin.PluginID(req.Id))
	if err != nil {
		if errors.Is(err, pluginsreg.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "")
		}
	}

	return &pbsvc.GetResponse{Plugin: p.PB()}, nil
}

func (s *Svc) Register(ctx context.Context, req *pbsvc.RegisterRequest) (*pbsvc.RegisterResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	settings, err := apiplugin.PluginSettingsFromProto(req.Settings)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "settings: %v", err)
	}

	if err := s.Registry.RegisterExternalPlugin(ctx, apiplugin.PluginID(req.Id), settings); err != nil {
		return nil, status.Errorf(codes.Unknown, "store: %v", err)
	}

	return &pbsvc.RegisterResponse{}, nil
}
