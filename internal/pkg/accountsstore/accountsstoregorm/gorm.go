package accountsstoregorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
)

type Store struct{ DB *gorm.DB }

var _ accountsstore.Store = &Store{}

func (db *Store) Create(ctx context.Context, name apiaccount.AccountName, data *apiaccount.AccountSettings) error {
	memo, err := marshalMemo(data.Memo())
	if err != nil {
		return err
	}

	a := account{
		Name:      name.String(),
		CreatedAt: time.Now(),
		Enabled:   data.Enabled(),
		Memo:      memo,
	}

	db_ := db.DB.
		WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&a)

	if err := db_.Error; err != nil {
		return fmt.Errorf("create: %w", err)
	}

	if db_.RowsAffected == 0 {
		return accountsstore.ErrAlreadyExists
	}

	return nil
}

func (db *Store) get(ctx context.Context, name apiaccount.AccountName) (*account, error) {
	var a account

	err := db.DB.
		WithContext(ctx).
		First(&a, "name = ?", name.String()).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, accountsstore.ErrNotFound
		}

		return nil, err
	}

	return &a, nil
}

func (db *Store) Get(ctx context.Context, name apiaccount.AccountName) (*apiaccount.Account, error) {
	a, err := db.get(ctx, name)
	if err != nil {
		return nil, err
	}

	return decodeAccount(a)
}

func (db *Store) BatchGet(ctx context.Context, names []apiaccount.AccountName) (map[apiaccount.AccountName]*apiaccount.Account, error) {
	var as []account

	// TODO: some sql magic to get only top versions.
	err := db.DB.
		WithContext(ctx).
		Find(&as, "name in ?", names).
		Error
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	m := make(map[apiaccount.AccountName]*apiaccount.Account, len(names))
	for _, a := range as {
		aid := apiaccount.AccountName(a.Name)

		if m[aid], err = decodeAccount(&a); err != nil {
			return nil, fmt.Errorf("decode account %q: %w", a.Name, err)
		}
	}

	for _, name := range names {
		if m[name] == nil {
			m[name] = nil
		}
	}

	return m, nil
}

func (db *Store) Update(ctx context.Context, name apiaccount.AccountName, data *apiaccount.AccountSettings) error {
	a, err := db.get(ctx, name)
	if err != nil {
		return err
	}

	if data == nil {
		return nil
	}

	a.Enabled = data.Enabled()
	a.UpdatedAt = time.Now()

	if memo := data.Memo(); len(memo) != 0 {
		if a.Memo, err = marshalMemo(memo); err != nil {
			return err
		}
	}

	if err := db.DB.WithContext(ctx).Save(a).Error; err != nil {
		// TODO: gracefully handle concurrent updates. (conflict on version)
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

func (db *Store) Setup(ctx context.Context) error {
	if err := db.DB.WithContext(ctx).AutoMigrate(&account{}); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}

	return nil
}

func (db *Store) Teardown(ctx context.Context) error {
	if err := db.DB.WithContext(ctx).Migrator().DropTable(&account{}); err != nil {
		return fmt.Errorf("drop: %w", err)
	}

	return nil
}
