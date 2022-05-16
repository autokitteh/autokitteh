package langtools

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	T "gitlab.com/softkitteh/autokitteh/cmd/ak/clitools"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang"
	_ "gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langall"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langrun"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langrun/grpclangrun"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langrun/locallangrun"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang/langtools"
)

var Settings struct {
	Catalog lang.Catalog
	Runs    langrun.Runs
}

func Catalog() lang.Catalog { return Settings.Catalog }
func Runs() langrun.Runs    { return Settings.Runs }

func Init() error {
	if addr := T.Addr(); addr != "builtin" && addr != "" {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			return fmt.Errorf("catalog grpc dial: %w", err)
		}

		Settings.Catalog, err = langtools.NewGRPCCatalogFromConn(T.Context, T.L().Named("grpccatalog"), conn)
		if err != nil {
			return fmt.Errorf("new grpc catalog: %w", err)
		}

		Settings.Runs = grpclangrun.NewRunsFromConn(T.L().Named("runs"), conn)
	} else {
		langtools.PermissiveCatalog.L.Set(T.L().Named("localcatalog"))

		Settings.Catalog = langtools.PermissiveCatalog

		Settings.Runs = locallangrun.NewRuns(T.L().Named("runs"), langtools.PermissiveCatalog, nil, nil)
	}

	return nil
}
