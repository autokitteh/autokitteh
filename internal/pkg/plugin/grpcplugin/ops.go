package grpcplugin

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autokitteh/autokitteh/gen/proto/stubs/go/pluginsprovidersvc"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiplugin"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/plugin"
)

func check(resp interface{ Validate() error }, err error) error {
	if err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.NotFound {
			return plugin.ErrPluginNotFound
		}

		return err
	}

	if err := resp.Validate(); err != nil {
		return err
	}

	return nil
}

func (p *GRPCPlugin) Describe(ctx context.Context) (*apiplugin.PluginDesc, error) {
	resp, err := p.Client.Describe(ctx, &pb.DescribeRequest{Id: string(p.ID)})
	if err := check(resp, err); err != nil {
		return nil, err
	}

	desc, err := apiplugin.PluginDescFromProto(resp.Desc)
	if err != nil {
		return nil, fmt.Errorf("descriptor: %w", err)
	}

	return desc, nil
}

func (p *GRPCPlugin) get(ctx context.Context, names []string) (map[string]*apivalues.Value, error) {
	resp, err := p.Client.GetValues(ctx, &pb.GetValuesRequest{Id: string(p.ID), Names: names})
	if err := check(resp, err); err != nil {
		return nil, err
	}

	vs := make(map[string]*apivalues.Value, len(resp.Values))
	for k, v := range resp.Values {
		var err error
		if vs[k], err = apivalues.ValueFromProto(v); err != nil {
			return nil, fmt.Errorf("value %q: %w", k, err)
		}
	}

	return vs, nil
}

func (p *GRPCPlugin) GetAll(ctx context.Context) (map[string]*apivalues.Value, error) {
	return p.get(ctx, nil)
}

func (p *GRPCPlugin) Get(ctx context.Context, name string) (*apivalues.Value, error) {
	vs, err := p.get(ctx, []string{name})
	if err != nil {
		return nil, err
	}

	return vs[name], nil
}

func (p *GRPCPlugin) Call(ctx context.Context, v *apivalues.Value, args []*apivalues.Value, kwargs map[string]*apivalues.Value) (*apivalues.Value, error) {
	resp, err := p.Client.CallValue(ctx, &pb.CallValueRequest{
		Id:     string(p.ID),
		Value:  v.PB(),
		Args:   apivalues.ValuesListToProto(args),
		Kwargs: apivalues.StringValueMapToProto(kwargs),
	})
	if err := check(resp, err); err != nil {
		return nil, err
	}

	switch ret := resp.Ret.(type) {
	case *pb.CallValueResponse_Error:
		perr, err := apiprogram.ErrorFromProto(ret.Error)
		if err != nil {
			return nil, fmt.Errorf("call return error parse: %w", err)
		}
		return nil, perr
	case *pb.CallValueResponse_Retval:
		// TODO: what happens with call values?
		return apivalues.ValueFromProto(ret.Retval)
	default:
		p.L.Error("unrecognized grpc call return value", "resp", resp)
		return nil, fmt.Errorf("unrecognized grpc call return value")
	}
}
