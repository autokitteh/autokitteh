package langgrpcsvc

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pblangsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/langsvc"
	pbprogram "github.com/autokitteh/autokitteh/api/gen/stubs/go/program"

	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langtools"
	"github.com/autokitteh/L"

	_ "github.com/autokitteh/autokitteh/internal/pkg/lang/langall"
)

type Svc struct {
	pblangsvc.UnimplementedLangServer

	L L.Nullable

	Catalog lang.Catalog
}

var _ pblangsvc.LangServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pblangsvc.RegisterLangServer(srv, s)

	if gw != nil {
		if err := pblangsvc.RegisterLangHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) ListLangs(context.Context, *pblangsvc.ListLangsRequest) (*pblangsvc.ListLangsResponse, error) {
	ls := s.Catalog.List()

	m := make(map[string]*pblangsvc.CatalogLang, len(ls))
	for l, exts := range ls {
		m[l] = &pblangsvc.CatalogLang{Exts: exts}
	}

	return &pblangsvc.ListLangsResponse{Langs: m}, nil
}

func (s *Svc) CompileModule(ctx context.Context, req *pblangsvc.CompileModuleRequest) (*pblangsvc.CompileModuleResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	path, err := apiprogram.PathFromProto(req.Path)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "path: %v", err)
	}

	mod, _, err := langtools.CompileModule(ctx, s.Catalog, req.Predecls, path, req.Src)
	if err != nil {
		st := status.Newf(codes.Unknown, "compile: %v", err)

		if perr := (&apiprogram.Error{}); errors.As(err, &perr) {
			d, err := st.WithDetails(perr.PB())
			if err != nil {
				s.L.Error("unable to add details to grpc error status", "err", err)
			} else {
				st = d
			}
		}

		return nil, st.Err()
	}

	var paths []*apiprogram.Path

	if req.GetDeps {
		paths, err = langtools.GetModuleDependencies(ctx, s.Catalog, mod)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "deps: %v", err)
		}
	}

	pbdeps := make([]*pbprogram.Path, len(paths))
	for i, path := range paths {
		pbdeps[i] = path.PB()
	}

	return &pblangsvc.CompileModuleResponse{
		Module: mod.PB(),
		Deps:   &pblangsvc.Dependencies{Ready: pbdeps},
	}, nil
}

func (s *Svc) IsCompilerVersionSupported(ctx context.Context, req *pblangsvc.IsCompilerVersionSupportedRequest) (*pblangsvc.IsCompilerVersionSupportedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	supported, err := langtools.IsCompilerVersionSupported(ctx, s.Catalog, req.Lang, req.Ver)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "%v", err)
	}

	return &pblangsvc.IsCompilerVersionSupportedResponse{Supported: supported}, nil
}

func (s *Svc) GetModuleDependencies(ctx context.Context, req *pblangsvc.GetModuleDependenciesRequest) (*pblangsvc.GetModuleDependenciesResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	mod, err := apiprogram.ModuleFromProto(req.Module)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid module: %v", err)
	}

	paths, err := langtools.GetModuleDependencies(ctx, s.Catalog, mod)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	return &pblangsvc.GetModuleDependenciesResponse{
		Deps: &pblangsvc.Dependencies{
			Ready: pathsToPB(paths),
		},
	}, nil
}

func pathsToPB(in []*apiprogram.Path) []*pbprogram.Path {
	l := make([]*pbprogram.Path, len(in))
	for i, p := range in {
		l[i] = p.PB()
	}
	return l
}
