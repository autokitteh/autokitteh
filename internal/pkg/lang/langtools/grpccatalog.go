package langtools

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	pblangsvc "github.com/autokitteh/autokitteh/gen/proto/stubs/go/langsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/internal/pkg/lang/langgrpc"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type grpcCatalog struct{ local LocalCatalog }

var _ lang.Catalog = &grpcCatalog{}

func (*grpcCatalog) Register(string, lang.CatalogLang) { panic("cannot register on a remote catalog") }

func (c *grpcCatalog) List() map[string][]string { return c.local.List() }

func (c *grpcCatalog) Acquire(name, scope string) (lang.Lang, error) {
	return c.local.Acquire(name, scope)
}

func NewGRPCCatalog(ctx context.Context, l L.L, client pblangsvc.LangClient, runClient pblangsvc.LangRunClient) (lang.Catalog, error) {
	resp, err := client.ListLangs(ctx, &pblangsvc.ListLangsRequest{})
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	cat := LocalCatalog{L: L.N(l)}

	for name, clang := range resp.Langs {
		l.Debug("registering language", "lang", name, "exts", clang.Exts)

		cat.Register(name, lang.CatalogLang{
			New: func(l L.L, name string) (lang.Lang, error) {
				return langgrpc.New(l.Named(name), name, client, runClient)
			},
			Exts: clang.Exts,
		})
	}

	return &grpcCatalog{local: cat}, nil
}

func NewGRPCCatalogFromConn(ctx context.Context, l L.L, conn *grpc.ClientConn) (lang.Catalog, error) {
	client, runClient := pblangsvc.NewLangClient(conn), pblangsvc.NewLangRunClient(conn)

	return NewGRPCCatalog(ctx, l, client, runClient)
}
