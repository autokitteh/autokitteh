package manifest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"go.autokitteh.dev/sdk/api/apiaccount"
)

type Account struct {
	Name     string            `json:"name"`
	Disabled bool              `json:"disabled"`
	Memo     map[string]string `json:"memo"`
}

func (a Account) API(name string) (*apiaccount.Account, error) {
	if a.Name != "" {
		name = a.Name
	}

	return apiaccount.NewAccount(
		apiaccount.AccountName(name),
		(&apiaccount.AccountSettings{}).
			SetEnabled(!a.Disabled).
			SetMemo(a.Memo),
		time.Now(),
		nil,
	)
}

func (a Account) Compile(name string) ([]*Action, error) {
	api, err := a.API(name)
	if err != nil {
		return nil, fmt.Errorf("invalid account: %w", err)
	}

	return []*Action{{
		Desc: fmt.Sprintf("create account %q", api.Name()),
		Run: func(ctx context.Context, env *Env) (string, error) {
			if env.Accounts == nil {
				return "", fmt.Errorf("have no accounts access")
			}

			err := env.Accounts.Create(ctx, api.Name(), api.Settings())
			if err != nil {
				if errors.Is(err, accountsstore.ErrAlreadyExists) {
					return "already exists", nil
				}

				return "failed", err
			}

			return "created", err
		},
	}}, nil
}
