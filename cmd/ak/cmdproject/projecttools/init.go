package projecttools

import (
	"fmt"

	T "github.com/autokitteh/autokitteh/cmd/ak/clitools"
	A "github.com/autokitteh/autokitteh/cmd/ak/cmdaccount/accounttools"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore/projectsstorefactory"
)

var Settings struct {
	projects projectsstore.Store
}

func Projects() projectsstore.Store { return Settings.projects }

func Init(spec string) error {
	if spec == "" {
		if addr := T.Addr(); addr != "" && addr != "builtin" {
			spec = fmt.Sprintf("grpc:%s", addr)
		}
	}

	if err := A.Init(spec); err != nil {
		return fmt.Errorf("accounts: %w", err)
	}

	var err error
	if Settings.projects, err = projectsstorefactory.OpenString(T.Context, T.L().Named("projectsstore"), spec, A.Accounts()); err != nil {
		return fmt.Errorf("projects: %w", err)
	}

	if err := Settings.projects.Setup(T.Context); err != nil {
		return fmt.Errorf("projects setup: %w", err)
	}

	return nil
}
