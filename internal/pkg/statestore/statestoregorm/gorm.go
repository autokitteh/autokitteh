package statestoregorm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/statestore"
)

// TODO: this all saves everything in a single row for each name.
//       should do it more redis-style to be more efficient and scalable.
// TODO: also implement for redis.

type Store struct {
	DB *gorm.DB
}

var _ statestore.Store = &Store{}

func set(ctx context.Context, db *gorm.DB, pid apiproject.ProjectID, name string, v *apivalues.Value) error {
	var j []byte

	if v != nil {
		if v.IsEphemeral() {
			return fmt.Errorf("cannot persist ephemeral values")
		}

		var err error
		if j, err = json.Marshal(v); err != nil {
			return fmt.Errorf("value marshal: %w", err)
		}
	}

	if err := db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(value{
		ProjectID: pid.String(),
		Name:      name,
		Value:     j,
		Metadata: statestore.Metadata{
			UpdatedAt: time.Now(),
		},
	}).Error; err != nil {
		return fmt.Errorf("create: %w", err)
	}

	return nil
}

func (s *Store) Set(ctx context.Context, pid apiproject.ProjectID, name string, v *apivalues.Value) error {
	return set(ctx, s.DB, pid, name, v)
}

func get(ctx context.Context, db *gorm.DB, pid apiproject.ProjectID, name string) (*apivalues.Value, *statestore.Metadata, error) {
	var v value

	if err := db.
		WithContext(ctx).
		First(&v, "project_id = ? AND name = ?", pid.String(), name).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, statestore.ErrNotFound
		}

		return nil, nil, fmt.Errorf("get: %w", err)
	}

	if v.Value == nil {
		return nil, nil, statestore.ErrNotFound
	}

	var vv apivalues.Value

	if err := json.Unmarshal(v.Value, &vv); err != nil {
		return nil, nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &vv, &v.Metadata, nil
}

func (s *Store) Get(ctx context.Context, pid apiproject.ProjectID, name string) (*apivalues.Value, *statestore.Metadata, error) {
	return get(ctx, s.DB, pid, name)
}

func (s *Store) List(ctx context.Context, pid apiproject.ProjectID) ([]string, error) {
	var ns []string

	if err := s.DB.WithContext(ctx).Model(&value{}).Where("project_id = ?", pid.String()).Pluck("name", &ns).Error; err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	return ns, nil
}

func (s *Store) Inc(ctx context.Context, pid apiproject.ProjectID, name string, amount int64) (ret *apivalues.Value, err error) {
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		old, _, err := get(ctx, tx, pid, name)
		if errors.Is(err, statestore.ErrNotFound) {
			old = apivalues.Integer(0)
		}

		if ret, err = apivalues.Inc(old, amount); err != nil {
			return err
		}

		return set(ctx, tx, pid, name, ret)
	})

	return
}

func (s *Store) Insert(ctx context.Context, pid apiproject.ProjectID, name string, idx int, v *apivalues.Value) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		old, _, err := get(ctx, tx, pid, name)
		if errors.Is(err, statestore.ErrNotFound) {
			if idx > 0 {
				return fmt.Errorf("variable not found and index > 0")
			}

			old = apivalues.List()
		}

		next, err := apivalues.Insert(old, idx, v)
		if err != nil {
			return err
		}

		return set(ctx, tx, pid, name, next)
	})
}

func (s *Store) Take(ctx context.Context, pid apiproject.ProjectID, name string, idx, count int) (taken *apivalues.Value, err error) {
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		old, _, err := get(ctx, tx, pid, name)
		if errors.Is(err, statestore.ErrNotFound) {
			old = apivalues.List()
		}

		next, takenList, err := apivalues.Take(old, idx, count)
		if err != nil {
			return err
		}

		if err := set(ctx, tx, pid, name, next); err != nil {
			return err
		}

		taken = apivalues.List(takenList...)

		return nil
	})

	return
}

func (s *Store) Index(ctx context.Context, pid apiproject.ProjectID, name string, idx int) (*apivalues.Value, error) {
	v, _, err := get(ctx, s.DB, pid, name)
	if err != nil {
		return nil, err
	}

	return apivalues.Index(v, idx)
}

func (s *Store) Length(ctx context.Context, pid apiproject.ProjectID, name string) (int, error) {
	v, _, err := get(ctx, s.DB, pid, name)
	if err != nil {
		return 0, err
	}

	return apivalues.Length(v)
}

func (s *Store) SetKey(ctx context.Context, pid apiproject.ProjectID, name string, k, v *apivalues.Value) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		old, _, err := get(ctx, tx, pid, name)
		if errors.Is(err, statestore.ErrNotFound) {
			old = apivalues.Dict()
		}

		next, err := apivalues.SetKey(old, k, v)
		if err != nil {
			return err
		}

		return set(ctx, tx, pid, name, next)
	})
}

func (s *Store) GetKey(ctx context.Context, pid apiproject.ProjectID, name string, k *apivalues.Value) (*apivalues.Value, error) {
	v, _, err := get(ctx, s.DB, pid, name)
	if err != nil {
		return nil, err
	}

	return apivalues.GetKey(v, k)
}

func (s *Store) Keys(ctx context.Context, pid apiproject.ProjectID, name string) (*apivalues.Value, error) {
	v, _, err := get(ctx, s.DB, pid, name)
	if err != nil {
		return nil, err
	}

	return apivalues.Keys(v)
}

func (s *Store) Setup(ctx context.Context) error {
	if err := s.DB.WithContext(ctx).AutoMigrate(&value{}); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}

	return nil
}

func (s *Store) Teardown(ctx context.Context) error {
	if err := s.DB.WithContext(ctx).Migrator().DropTable(&value{}); err != nil {
		return fmt.Errorf("drop: %w", err)
	}

	return nil
}
