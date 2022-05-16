package pluginsgrpcsvc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbsvc "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/pluginsprovidersvc"
	pbvalues "gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go/values"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/plugin"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiplugin"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
	L "gitlab.com/softkitteh/autokitteh/pkg/l"
)

type Svc struct {
	pbsvc.UnimplementedPluginsProviderServer

	Plugins map[apiplugin.PluginID]plugin.Plugin

	L L.Nullable
}

var _ pbsvc.PluginsProviderServer = &Svc{}

func (s *Svc) Register(srv *grpc.Server) { pbsvc.RegisterPluginsProviderServer(srv, s) }

func (s *Svc) List(ctx context.Context, req *pbsvc.ListRequest) (*pbsvc.ListResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	ids := make([]string, 0, len(s.Plugins))
	for plid := range s.Plugins {
		ids = append(ids, plid.String())
	}

	return &pbsvc.ListResponse{Ids: ids}, nil
}

func (s *Svc) Describe(ctx context.Context, req *pbsvc.DescribeRequest) (*pbsvc.DescribeResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	pl, ok := s.Plugins[apiplugin.PluginID(req.Id)]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "not found")
	}

	desc, err := pl.Describe(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "describe: %v", err)
	}

	return &pbsvc.DescribeResponse{Desc: desc.PB()}, nil
}

func (s *Svc) CallValue(ctx context.Context, req *pbsvc.CallValueRequest) (*pbsvc.CallValueResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	pl, ok := s.Plugins[apiplugin.PluginID(req.Id)]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "plugin not found")
	}

	kwargs, err := apivalues.StringValueMapFromProto(req.Kwargs)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "kwargs: %v", err)
	}

	args, err := apivalues.ValuesListFromProto(req.Args)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "args: %v", err)
	}

	v, err := apivalues.ValueFromProto(req.Value)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "value: %v", err)
	}

	ret, err := pl.Call(
		ctx,
		v,
		args,
		kwargs,
	)
	if err != nil {
		return &pbsvc.CallValueResponse{Ret: &pbsvc.CallValueResponse_Error{Error: apiprogram.ImportError(err).PB()}}, nil
	}

	return &pbsvc.CallValueResponse{Ret: &pbsvc.CallValueResponse_Retval{Retval: ret.PB()}}, nil
}

func (s *Svc) GetValues(ctx context.Context, req *pbsvc.GetValuesRequest) (*pbsvc.GetValuesResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	pl, ok := s.Plugins[apiplugin.PluginID(req.Id)]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "plugin not found")
	}

	var (
		vs  map[string]*apivalues.Value
		err error
	)

	if len(req.Names) == 0 {
		if vs, err = pl.GetAll(ctx); err != nil {
			return nil, status.Errorf(codes.Unknown, "get all: %v", err)
		}
	} else {
		vs = make(map[string]*apivalues.Value, len(req.Names))
		for _, name := range req.Names {
			if vs[name], err = pl.Get(ctx, name); err != nil {
				return nil, status.Errorf(codes.Unknown, "get %q: %v", name, err)
			}
		}
	}

	pbvs := make(map[string]*pbvalues.Value, len(vs))
	for k, v := range vs {
		pbvs[k] = v.PB()
	}

	return &pbsvc.GetValuesResponse{Values: pbvs}, nil
}
