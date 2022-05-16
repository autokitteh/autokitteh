package accounttools

import (
	"fmt"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore/accountsstorefactory"
)

var Settings struct {
	accounts accountsstore.Store
}

func Accounts() accountsstore.Store { return Settings.accounts }

func Init(spec string) error {
	addr := T.Addr()

	if spec == "" {
		if addr != "" && addr != "builtin" {
			spec = fmt.Sprintf("grpc:%s", addr)
		}
	}

	var err error
	if Settings.accounts, err = accountsstorefactory.OpenString(T.Context, T.L().Named("accountsstore"), spec); err != nil {
		return fmt.Errorf("accountsstore: %w", err)
	}

	if err := Settings.accounts.Setup(T.Context); err != nil {
		return fmt.Errorf("accountsstore setup: %w", err)
	}

	return nil
}
