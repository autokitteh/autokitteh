package githubinstalls

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v42/github"

	"github.com/autokitteh/stores/kvstore"
)

type Config struct {
	AppID         int64  `envconfig:"APP_ID" json:"app_id"`
	AppPrivateKey string `envconfig:"APP_PRIVATE_KEY" json:"app_private_key"`
}

type Installs struct {
	Config Config
	Store  kvstore.Store
}

//=======
// TODO: remove this and properly pass installs to wherever needed.

var insts *Installs

func New(cfg Config, store kvstore.Store) *Installs {
	insts = &Installs{Config: cfg, Store: store}
	return insts
}

func GetInstalls() *Installs { return insts }

//=======

func (i *Installs) Add(ctx context.Context, owner, repo string, inst *github.Installation) error {
	if i == nil {
		return fmt.Errorf("installations disabled")
	}

	return kvstore.PutJSON(ctx, i.Store, fmt.Sprintf("%s:%s", owner, repo), inst)
}

func (i *Installs) GetClient(ctx context.Context, owner, repo string) (*github.Client, error) {
	if i == nil {
		return nil, nil
	}

	var inst github.Installation

	if err := kvstore.GetJSON(ctx, i.Store, fmt.Sprintf("%s:%s", owner, repo), &inst); err != nil {
		if errors.Is(err, kvstore.ErrNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("get install error: %w", err)
	}

	if inst.ID == nil {
		return nil, fmt.Errorf("invalid installation, missing id")
	}

	tr := http.DefaultTransport
	itr, err := ghinstallation.New(tr, i.Config.AppID, *inst.ID, []byte(i.Config.AppPrivateKey))
	if err != nil {
		return nil, fmt.Errorf("new client install: %w", err)
	}

	return github.NewClient(&http.Client{Transport: itr}), nil
}
