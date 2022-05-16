package projectsstoregorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gitlab.com/softkitteh/autokitteh/internal/pkg/accountsstore"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/projectsstore"
)

type Store struct {
	DB            *gorm.DB
	AccountsStore accountsstore.Store
	NoAccounts    bool
}

var _ projectsstore.Store = &Store{}

func (db *Store) Create(
	ctx context.Context,
	aname apiaccount.AccountName,
	id apiproject.ProjectID,
	d *apiproject.ProjectSettings,
) (apiproject.ProjectID, error) {
	memo, err := marshalMemo(d.Memo())
	if err != nil {
		return "", err
	}

	predecls, err := marshalPredecls(d.Predecls())
	if err != nil {
		return "", err
	}

	plugins, err := marshalPlugins(d.Plugins())
	if err != nil {
		return "", err
	}

	if !db.NoAccounts {
		if db.AccountsStore == nil {
			return "", fmt.Errorf("not set up with accounts database")
		}

		if _, err := db.AccountsStore.Get(ctx, aname); err != nil {
			if errors.Is(err, accountsstore.ErrNotFound) {
				return "", fmt.Errorf("%w: account not found", projectsstore.ErrInvalidAccount)
			}

			return "", fmt.Errorf("get account: %w", err)
		}
	}

	if id == projectsstore.AutoProjectID {
		id = apiproject.NewProjectID()
	}

	p := project{
		ID:          id.String(),
		AccountName: aname.String(),
		CreatedAt:   time.Now(),
		Enabled:     d.Enabled(),
		MainPath:    d.MainPath().String(),
		Predecls:    predecls,
		Plugins:     plugins,
		Name:        d.Name(),
		Memo:        memo,
	}

	db_ := db.DB.
		WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&p)

	if err := db_.Error; err != nil {
		return "", fmt.Errorf("create: %w", err)
	}

	if db_.RowsAffected == 0 {
		return "", projectsstore.ErrAlreadyExists
	}

	return id, nil
}

func (db *Store) get(ctx context.Context, id apiproject.ProjectID) (*project, error) {
	var p project

	err := db.DB.
		WithContext(ctx).
		First(&p, "id = ?", id.String()).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, projectsstore.ErrNotFound
		}

		return nil, err
	}

	return &p, nil
}

func (db *Store) Get(ctx context.Context, id apiproject.ProjectID) (*apiproject.Project, error) {
	p, err := db.get(ctx, id)
	if err != nil {
		return nil, err
	}

	return decodeProject(p)
}

func (db *Store) BatchGet(ctx context.Context, ids []apiproject.ProjectID) (map[apiproject.ProjectID]*apiproject.Project, error) {
	var ps []project

	// TODO: some sql magic to get only top versions.
	err := db.DB.
		WithContext(ctx).
		Find(&ps, "id in ?", ids).
		Error
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	m := make(map[apiproject.ProjectID]*apiproject.Project, len(ids))
	for _, p := range ps {
		if m[apiproject.ProjectID(p.ID)], err = decodeProject(&p); err != nil {
			return nil, fmt.Errorf("decode project %q: %w", p.ID, err)
		}
	}

	for _, id := range ids {
		if m[id] == nil {
			m[id] = nil
		}
	}

	return m, nil
}

func (db *Store) Update(
	ctx context.Context,
	id apiproject.ProjectID,
	d *apiproject.ProjectSettings,
) error {
	p, err := db.get(ctx, id)
	if err != nil {
		return err
	}

	p.Enabled = d.Enabled()

	if d.MainPath() != nil {
		p.MainPath = d.MainPath().String()
	}

	p.UpdatedAt = time.Now()

	if memo := d.Memo(); len(memo) != 0 {
		if p.Memo, err = marshalMemo(memo); err != nil {
			return err
		}
	}

	if n := d.Name(); n != "" {
		p.Name = n
	}

	if err := db.DB.WithContext(ctx).Save(p).Error; err != nil {
		// TODO: gracefully handle concurrent updates. (conflict on version)
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

func (db *Store) Setup(ctx context.Context) error {
	if err := db.DB.WithContext(ctx).AutoMigrate(&project{}); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}

	return nil
}

func (db *Store) Teardown(ctx context.Context) error {
	if err := db.DB.WithContext(ctx).Migrator().DropTable(&project{}); err != nil {
		return fmt.Errorf("drop: %w", err)
	}

	return nil
}
