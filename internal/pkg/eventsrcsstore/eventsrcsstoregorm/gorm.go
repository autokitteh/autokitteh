package eventsrcsstoregorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsrcsstore"
)

type Store struct{ DB *gorm.DB }

var _ eventsrcsstore.Store = &Store{}

func (s *Store) Add(ctx context.Context, id apieventsrc.EventSourceID, data *apieventsrc.EventSourceSettings) error {
	r := eventsrc{
		SrcID:       id.String(),
		AccountName: id.AccountName().String(),
		Enabled:     data.Enabled(),
		CreatedAt:   time.Now(),
	}

	db_ := s.DB.
		WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&r)

	if err := db_.Error; err != nil {
		return fmt.Errorf("create: %w", err)
	}

	if db_.RowsAffected == 0 {
		return eventsrcsstore.ErrAlreadyExists
	}

	return nil
}

func (s *Store) get(ctx context.Context, id apieventsrc.EventSourceID) (*eventsrc, error) {
	var r eventsrc

	err := s.DB.
		WithContext(ctx).
		Where("src_id = ?", id.String()).
		First(&r).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, eventsrcsstore.ErrNotFound
		}

		return nil, err
	}

	return &r, nil
}

func (s *Store) Get(ctx context.Context, id apieventsrc.EventSourceID) (*apieventsrc.EventSource, error) {
	r, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}

	src, err := decodeEventSource(r)
	if err != nil {
		return nil, fmt.Errorf("src: %w", err)
	}

	return src, nil
}

func (s *Store) Update(ctx context.Context, id apieventsrc.EventSourceID, data *apieventsrc.EventSourceSettings) error {
	if data == nil {
		return nil
	}

	r, err := s.get(ctx, id)
	if err != nil {
		return err
	}

	r.Enabled = data.Enabled()
	r.UpdatedAt = time.Now()

	if err := s.DB.WithContext(ctx).Save(&r).Error; err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

func (s *Store) List(ctx context.Context, aname *apiaccount.AccountName) ([]apieventsrc.EventSourceID, error) {
	var srcs []eventsrc

	if aname == nil {
		// TODO
		return nil, fmt.Errorf("global list not supported yet")
	}

	if err := s.DB.WithContext(ctx).
		Model(&eventsrc{}).
		Where("account_id = ?", aname.String()).
		Distinct().
		Select("src_id").
		Find(&srcs).
		Error; err != nil {
		return nil, fmt.Errorf("pluck: %w", err)
	}

	ids := make([]apieventsrc.EventSourceID, len(srcs))
	for i, src := range srcs {
		ids[i] = apieventsrc.EventSourceID(src.SrcID)
	}

	return ids, nil
}

func (s *Store) AddProjectBinding(ctx context.Context, srcid apieventsrc.EventSourceID, pid apiproject.ProjectID, name, assoc, cfg string, approved bool, data *apieventsrc.EventSourceProjectBindingSettings) error {
	r := binding{
		SrcID:        srcid.String(),
		Name:         name,
		ProjectID:    pid.String(),
		Enabled:      data.Enabled(),
		Association:  assoc,
		SourceConfig: cfg,
		Approved:     approved,
	}

	db_ := s.DB.
		WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&r)

	if err := db_.Error; err != nil {
		return fmt.Errorf("create: %w", err)
	}

	if db_.RowsAffected == 0 {
		return eventsrcsstore.ErrAlreadyExists
	}

	return nil
}

func (s *Store) UpdateProjectBinding(ctx context.Context, srcid apieventsrc.EventSourceID, pid apiproject.ProjectID, name string, approved bool, data *apieventsrc.EventSourceProjectBindingSettings) error {
	if data == nil {
		return nil
	}

	var r binding

	err := s.DB.
		WithContext(ctx).
		Where("project_id = ?", pid.String()).
		Where("src_id = ?", srcid.String()).
		Where("name = ?", name).
		First(&r).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return eventsrcsstore.ErrNotFound
		}

		return err
	}

	r.Approved = approved
	r.Enabled = data.Enabled()
	r.UpdatedAt = time.Now()

	if err := s.DB.WithContext(ctx).Save(&r).Error; err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

func (s *Store) GetProjectBindings(ctx context.Context, srcid *apieventsrc.EventSourceID, pid *apiproject.ProjectID, name, assoc string, onlyApproved bool) ([]*apieventsrc.EventSourceProjectBinding, error) {
	if pid == nil && srcid == nil {
		// TODO: once paginh is supported, allow.
		return nil, fmt.Errorf("either source id or project id must be specified")
	}

	var rs []binding

	q := s.DB.WithContext(ctx).Order("created_at asc")

	if pid != nil {
		q = q.Where("project_id = ?", pid.String())
	}

	if srcid != nil {
		q = q.Where("src_id = ?", srcid.String())
	}

	if name != "" {
		q = q.Where("name = ?", name)
	}

	if assoc != "" {
		q = q.Where("association = ?", assoc)
	}

	if onlyApproved {
		q = q.Where("approved = ?", true)
	}

	if err := q.Find(&rs).Error; err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	bs := make([]*apieventsrc.EventSourceProjectBinding, 0, len(rs))
	for _, r := range rs {
		b, err := decodeBinding(&r)
		if err != nil {
			return nil, fmt.Errorf("record %v/%v: %w", r.SrcID, r.ProjectID, err)
		}

		bs = append(bs, b)
	}

	return bs, nil
}

func (s *Store) Setup(ctx context.Context) error {
	if err := s.DB.WithContext(ctx).AutoMigrate(&eventsrc{}, &binding{}); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}

	return nil
}

func (s *Store) Teardown(ctx context.Context) error {
	if err := s.DB.WithContext(ctx).Migrator().DropTable(&eventsrc{}, &binding{}); err != nil {
		return fmt.Errorf("drop: %w", err)
	}

	return nil
}
