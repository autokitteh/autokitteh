package accountsstore

import (
	"context"
	"errors"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiaccount"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Store interface {
	Create(context.Context, apiaccount.AccountName, *apiaccount.AccountSettings) error
	Update(context.Context, apiaccount.AccountName, *apiaccount.AccountSettings) error
	Get(context.Context, apiaccount.AccountName) (*apiaccount.Account, error)
	BatchGet(context.Context, []apiaccount.AccountName) (map[apiaccount.AccountName]*apiaccount.Account, error)
	Setup(context.Context) error
	Teardown(context.Context) error
}
