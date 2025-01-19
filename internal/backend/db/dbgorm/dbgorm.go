package dbgorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	_ "ariga.io/atlas-provider-gorm/gormschema"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"go.jetify.com/typeid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
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

func connect(_ context.Context, z *zap.Logger, cfg *Config) (*gorm.DB, error) {
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

func initGoose(client *sql.DB, dialect string) (ver int64, err error) {
	goose.SetBaseFS(migrations.Migrations)

	if err = goose.SetDialect(dialect); err != nil {
		return
	}

	if ver, err = goose.EnsureDBVersion(client); err != nil {
		err = fmt.Errorf("failed to ensure DB version: %w", err)
	}

	return
}

func (db *gormdb) Migrate(ctx context.Context) error {
	db.z.Info("migrating")

	client := db.client(true)

	ver, err := initGoose(client, db.cfg.Type)
	if err != nil {
		return err
	}

	migrationsDir := db.cfg.Type
	if err := goose.UpContext(ctx, client, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	if ver >= 20241225035414 {
		// This is needed only when initializing the database from a specific version.
		if err := db.backfillUsersAndOrgs(ctx); err != nil {
			return fmt.Errorf("backfill users and orgs: %w", err)
		}
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

	// Find all users that don't have default_org_id set. These are users that were created before the migration.
	var users []scheme.User
	if err := gdb.Where("default_org_id is NULL").Find(&users).Error; err != nil {
		return err
	}

	l.Info("backfilling users and orgs for users without default orgs", zap.Int("users", len(users)))

	usersMap := make(map[uuid.UUID]scheme.User, len(users))

	for i, user := range users {
		l := l.With(zap.String("user_id", user.UserID.String()), zap.Int("i", i))

		// Prepare a personal org for each user.
		org := scheme.Org{
			OrgID:       kittehs.Must1(uuid.NewV7()),
			DisplayName: user.DisplayName + "'s Personal Org",
			Base: scheme.Base{
				CreatedBy: authusers.SystemUser.DefaultOrgID().UUIDValue(),
				CreatedAt: kittehs.Now().UTC(),
			},
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

		// Register the user as its own org member.

		l.Info("registering user as org member")

		if err := gdb.Save(&scheme.OrgMember{
			OrgID:  org.OrgID,
			UserID: user.UserID,
			Base: scheme.Base{
				CreatedBy: authusers.SystemUser.DefaultOrgID().UUIDValue(),
				CreatedAt: kittehs.Now().UTC(),
			},
		}).Error; err != nil {
			return err
		}
	}

	// Backfill projects.

	var projects []scheme.Project

	// Find all projects that don't have org_id set. These are projects that were created before the migration.
	if err := gdb.Where("org_id is NULL").Find(&projects).Error; err != nil {
		return err
	}

	l.Info("backfilling projects without org", zap.Int("projects", len(projects)))

	for i, project := range projects {
		l := l.With(zap.String("project_id", project.ProjectID.String()), zap.Int("i", i))

		// Find the ownership of the project, which associates the project with the user that created it.
		var ownership scheme.Ownership
		if err := gdb.Where("entity_id = ?", project.ProjectID).First(&ownership).Error; err != nil {
			l.Error("failed to find ownership", zap.Error(err))
			continue
		}

		// The user id in the ownership can be either a UUID or a TypeID.
		uid, err := uuid.Parse(ownership.UserID)
		if err != nil {
			aid, err := typeid.Parse[typeid.AnyID](ownership.UserID)
			if err != nil {
				l.Error("failed to parse user id", zap.Error(err))
				continue
			}

			if uid, err = uuid.Parse(aid.UUID()); err != nil {
				l.Error("failed to parse user uuid", zap.Error(err))
				continue
			}
		}

		l = l.With(zap.String("user_id", uid.String()))

		var oid uuid.UUID

		if uid == authusers.DefaultUser.ID().UUIDValue() {
			// This is the default user, it already has a default org.
			oid = authusers.DefaultUser.DefaultOrgID().UUIDValue()
		} else {
			user, found := usersMap[uid]
			if !found {
				// User is not found since it already had a default org before migration, and just now updating its projects.
				if err := gdb.Where("user_id = ?", uid).First(&user).Error; err != nil {
					l.Error("failed to find user", zap.Error(err))
					continue
				}
			}

			oid = user.DefaultOrgID
		}

		l = l.With(zap.String("org_id", oid.String()))

		l.Info("updating project with org")

		// Associate the project with the user's default org (which is probably its personal org).
		err = gdb.Model(&scheme.Project{}).Where("project_id = ?", project.ProjectID).Update("org_id", oid).Error
		if err != nil {
			l.Error("failed to update project", zap.Error(err))
			continue
		}
	}

	return nil
}

func (db *gormdb) MigrationRequired(ctx context.Context) (bool, int64, error) {
	client := db.client(false)
	dbversion, err := initGoose(client, db.cfg.Type)
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
