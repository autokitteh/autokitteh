package dbgorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"go.jetify.com/typeid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "ariga.io/atlas-provider-gorm/gormschema"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/gormkitteh"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/migrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type Config = gormkitteh.Config

type gormdb struct {
	z   *zap.Logger
	db  *gorm.DB
	cfg *Config

	// See https://github.com/mattn/go-sqlite3/issues/274.
	// Used only for protecting writes when using sqlite.
	//
	// TODO(ENG-190): This is not... ideal. Really find a better
	//                solution. Everything else I tried didn't work.
	mu *sync.Mutex
}

var _ db.DB = (*gormdb)(nil)

func New(z *zap.Logger, cfg *Config) (db.DB, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	cfg, err := cfg.Explicit()
	if err != nil {
		return nil, err
	}

	db := &gormdb{z: z, cfg: cfg}

	if cfg.Type == "sqlite" {
		db.mu = new(sync.Mutex)
	}

	return db, nil
}

func (db *gormdb) GormDB() *gorm.DB { return db.db }

func connect(ctx context.Context, z *zap.Logger, cfg *Config) (*gorm.DB, error) {
	client, err := gormkitteh.OpenZ(z, cfg, func(cfg *gorm.Config) {
		cfg.SkipDefaultTransaction = true
	})
	if err != nil {
		return nil, fmt.Errorf("opendb: %w", err)
	}
	sqlDB, err := client.DB()
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	return client, nil
}

func (db *gormdb) Connect(ctx context.Context) (err error) {
	db.db, err = connect(ctx, db.z.Named("gorm"), db.cfg)
	return
}

func (db *gormdb) locked(f func(db *gormdb) error) error {
	if db.mu != nil {
		db.mu.Lock()
		defer db.mu.Unlock()
	}

	return f(db)
}

func translateError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows):
		return sdkerrors.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return sdkerrors.ErrAlreadyExists
	case errors.Is(err, sdkerrors.ErrAlreadyExists):
		return err
	case errors.Is(err, sdkerrors.ErrNotFound):
		return err
	default:
		return fmt.Errorf("db: %w", err)
	}
}

func gormErrNotFoundToForeignKey(err error) error {
	if err == gorm.ErrRecordNotFound {
		return gorm.ErrForeignKeyViolated
	}
	return err
}

var fkStmtByDB = map[string]map[bool]string{
	"sqlite": {
		true:  "PRAGMA foreign_keys = ON",
		false: "PRAGMA foreign_keys = OFF",
	},
	"postgres": {
		// in PG foreign keys implemented as triggers. Setting `session_replication_role'
		// to `replica' prevents firing triggers, thus effectively disables foreign keys
		true:  "SET session_replication_role = DEFAULT",
		false: "SET session_replication_role = replica",
	},
}

func foreignKeys(gormdb *gormdb, enable bool) {
	if _, found := fkStmtByDB[gormdb.cfg.Type]; !found {
		panic(fmt.Errorf("unknown DB type: %s", gormdb.cfg.Type))
	}
	stmt := fkStmtByDB[gormdb.cfg.Type][enable]
	gormdb.db.Exec(stmt)
}

func initGoose(client *sql.DB, dialect string) error {
	goose.SetBaseFS(migrations.Migrations)

	if err := goose.SetDialect(dialect); err != nil {
		return err
	}

	if _, err := goose.EnsureDBVersion(client); err != nil {
		return fmt.Errorf("failed to ensure DB version: %w", err)
	}
	return nil
}

func (db *gormdb) Migrate(ctx context.Context) error {
	db.z.Info("migrating")

	client := db.client(true)

	if err := initGoose(client, db.cfg.Type); err != nil {
		return err
	}

	migrationsDir := db.cfg.Type
	if err := goose.UpContext(ctx, client, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	if err := db.backfillUsersAndOrgs(ctx); err != nil {
		return fmt.Errorf("backfill users and orgs: %w", err)
	}

	return nil
}

func (db *gormdb) backfillUsersAndOrgs(ctx context.Context) error {
	l := db.z

	gdb, err := connect(ctx, db.z.Named("gorm-migrate"), db.cfg)
	if err != nil {
		return err
	}

	gdb = gdb.WithContext(ctx)

	// Backfill users.

	var users []scheme.User
	if err := gdb.Where("default_org_id is NULL").Find(&users).Error; err != nil {
		return err
	}

	l.Info("backfilling users and orgs for users without default orgs", zap.Int("users", len(users)))

	usersMap := make(map[uuid.UUID]scheme.User, len(users))

	for i, user := range users {
		l := l.With(zap.String("user_id", user.UserID.String()), zap.Int("i", i))

		org := scheme.Org{
			OrgID:       kittehs.Must1(uuid.NewV7()),
			DisplayName: fmt.Sprintf("%s's Personal Org", user.DisplayName),
		}

		l = l.With(zap.String("org_id", org.OrgID.String()))

		l.Info("creating org for user")

		if err := gdb.Create(&org).Error; err != nil {
			return err
		}

		l.Info("updating user with org")

		user.DefaultOrgID = org.OrgID
		if err := gdb.Save(&user).Error; err != nil {
			return err
		}

		usersMap[user.UserID] = user
	}

	// Backfill projects.

	var projects []struct {
		scheme.Project
		scheme.Ownership
	}

	if err := gdb.Where("org_id is NULL").Joins("left join ownerships on ownerships.entity_id = projects.project_id").Find(&projects).Error; err != nil {
		return err
	}

	l.Info("backfilling projects without org", zap.Int("projects", len(projects)))

	for i, project := range projects {
		l := l.With(zap.String("project_id", project.ProjectID.String()), zap.String("raw_user_id", project.UserID), zap.Int("i", i))

		uid, err := uuid.Parse(project.UserID)
		if err != nil {
			aid, err := typeid.Parse[typeid.AnyID](project.UserID)
			if err != nil {
				l.Error("failed to parse user id", zap.Error(err))
				continue
			}

			if uid, err = uuid.Parse(aid.UUID()); err != nil {
				l.Error("failed to parse user uuid", zap.Error(err))
				continue
			}
		}

		user, found := usersMap[uid]
		if !found {
			// user is not found since it already had a default org before migration, and just now updating its projects.
			if err := gdb.Where("user_id = ?", uid).First(&user).Error; err != nil {
				l.Error("failed to find user", zap.Error(err))
				continue
			}
		}

		l = l.With(zap.String("user_id", user.UserID.String()), zap.String("org_id", user.DefaultOrgID.String()))

		l.Info("updating project with org")

		err = gdb.Model(&scheme.Project{}).Update("org_id", user.DefaultOrgID).Where("project_id = ?", project.ProjectID).Error
		if err != nil {
			l.Error("failed to update project", zap.Error(err))
			continue
		}
	}

	return nil
}

func (db *gormdb) MigrationRequired(ctx context.Context) (bool, int64, error) {
	client := db.client(false)
	if err := initGoose(client, db.cfg.Type); err != nil {
		return false, 0, err
	}

	dbversion, err := goose.GetDBVersion(client)
	if err != nil {
		return false, 0, err
	}

	migrationsDir := db.cfg.Type
	requiredMigrations, err := goose.CollectMigrations(migrationsDir, dbversion, int64((1<<63)-1))
	if err != nil && !errors.Is(err, goose.ErrNoMigrationFiles) {
		return false, 0, err
	}

	return len(requiredMigrations) > 0, dbversion, nil
}

func (db *gormdb) migrate(ctx context.Context) error {
	required, dbVersion, err := db.MigrationRequired(ctx)
	if err != nil {
		return err
	}
	if !required {
		return nil
	}

	z := db.z.With(zap.Int64("db_version", dbVersion))

	z.Info("migration required")

	if db.cfg.AutoMigrate || dbVersion == 0 {
		return db.Migrate(ctx)
	}

	return errors.New("db migrations required") // TODO: maybe more details
}

func (db *gormdb) seed(ctx context.Context) error {
	if db.cfg.SeedCommands == "" {
		return nil
	}

	db.z.Info("seeding")

	cmd := db.db.WithContext(ctx).Debug().Exec(db.cfg.SeedCommands)

	db.z.Info("done seeding", zap.Int64("rows_affected", cmd.RowsAffected))

	return translateError(cmd.Error)
}

func (db *gormdb) Setup(ctx context.Context) error {
	isSqlite := db.cfg.Type == "sqlite"
	if isSqlite {
		foreignKeys(db, false)
		defer foreignKeys(db, true)
	}

	if err := db.migrate(ctx); err != nil {
		return err
	}

	if err := db.seed(ctx); err != nil {
		return err
	}

	return nil
}

// TODO: not sure this will work with the connect method
func (db *gormdb) Debug() db.DB {
	return &gormdb{
		z:  db.z,
		db: db.db.Debug(),
	}
}

// NOTE: no ctx is passed since in all places it's already applied
func getOne[T any](db *gorm.DB, where string, args ...any) (*T, error) {
	var r []T
	result := db.Where(where, args...).Limit(2).Find(&r)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if result.RowsAffected > 1 {
		return nil, gorm.ErrDuplicatedKey
	}
	return &r[0], nil
}

// TODO: this not working for deployments. Consider delete this function
func delete[T any](db *gorm.DB, ctx context.Context, where string, args ...any) error {
	var r T
	result := db.WithContext(ctx).Where(where, args...).Delete(&r)
	if result.Error != nil {
		return translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return sdkerrors.ErrNotFound
	}

	return nil
}

func (db *gormdb) client(debug bool) *sql.DB {
	q := db.db
	if debug {
		q = q.Debug()
	}

	return kittehs.Must1(q.DB())
}
