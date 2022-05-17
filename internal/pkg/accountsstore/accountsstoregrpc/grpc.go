package accountsstoregrpc

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbaccountsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/accountsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
)

type Store struct{ Client pbaccountsvc.AccountsClient }

var _ accountsstore.Store = &Store{}

func (as *Store) Create(ctx context.Context, name apiaccount.AccountName, d *apiaccount.AccountSettings) error {
	resp, err := as.Client.CreateAccount(
		ctx,
		&pbaccountsvc.CreateAccountRequest{Name: name.String(), Settings: d.PB()},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.AlreadyExists {
				return accountsstore.ErrAlreadyExists
			}
		}

		return fmt.Errorf("create: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return fmt.Errorf("resp validate: %w", err)
	}

	return nil
}

func (as *Store) Update(ctx context.Context, name apiaccount.AccountName, d *apiaccount.AccountSettings) error {
	resp, err := as.Client.UpdateAccount(
		ctx,
		&pbaccountsvc.UpdateAccountRequest{
			Name:     name.String(),
			Settings: d.PB(),
		},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return accountsstore.ErrNotFound
			}
		}

		return fmt.Errorf("update: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return fmt.Errorf("resp validate: %w", err)
	}

	return nil
}

func (as *Store) Get(ctx context.Context, name apiaccount.AccountName) (*apiaccount.Account, error) {
	resp, err := as.Client.GetAccount(
		ctx,
		&pbaccountsvc.GetAccountRequest{Name: name.String()},
	)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil, accountsstore.ErrNotFound
			}
		}

		return nil, fmt.Errorf("get: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp validate: %w", err)
	}

	a, err := apiaccount.AccountFromProto(resp.Account)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (db *Store) BatchGet(ctx context.Context, names []apiaccount.AccountName) (map[apiaccount.AccountName]*apiaccount.Account, error) {
	resp, err := db.Client.GetAccounts(
		ctx,
		&pbaccountsvc.GetAccountsRequest{
			Names: lo.Map(
				names,
				func(n apiaccount.AccountName, _ int) string { return n.String() },
			),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("resp validate: %w", err)
	}

	m := make(map[apiaccount.AccountName]*apiaccount.Account, len(names))

	for _, name := range names {
		m[name] = nil
	}

	for _, pbp := range resp.Accounts {
		if m[apiaccount.AccountName(pbp.Name)], err = apiaccount.AccountFromProto(pbp); err != nil {
			return nil, fmt.Errorf("invalid account %q: %w", pbp.Name, err)
		}
	}

	return m, nil
}

func (as *Store) Setup(ctx context.Context) error    { return fmt.Errorf("not supported through grpc") }
func (as *Store) Teardown(ctx context.Context) error { return fmt.Errorf("not supported through grpc") }
