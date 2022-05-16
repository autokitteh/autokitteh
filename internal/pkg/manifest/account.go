package manifest

import (
	"context"
	"fmt"
	"time"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
)

type Account struct {
	Name     string            `json:"name"`
	Disabled bool              `json:"disabled"`
	Memo     map[string]string `json:"memo"`
}

func (a Account) API() (*apiaccount.Account, error) {
	return apiaccount.NewAccount(
		apiaccount.AccountName(a.Name),
		(&apiaccount.AccountSettings{}).
			SetEnabled(!a.Disabled).
			SetMemo(a.Memo),
		time.Now(),
		nil,
	)
}

func (a Account) Compile() ([]*Action, error) {
	api, err := a.API()
	if err != nil {
		return nil, fmt.Errorf("invalid account: %w", err)
	}

	return []*Action{{
		Desc: fmt.Sprintf("create account %s", api.Name()),
		Run: func(ctx context.Context, env *Env) (string, error) {
			err := env.Accounts.Create(ctx, api.Name(), api.Settings())
			if err != nil {
				return "failed", err
			}

			return "created", err
		},
	}}, nil
}
