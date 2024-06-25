package dbgorm

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func userFromContext(ctx context.Context) (*scheme.User, error) {
	user := authcontext.GetAuthnUser(ctx)
	if !user.IsValid() {
		user = sdktypes.DefaultUser
	}

	data := user.Data()
	name := data["name"]
	email := data["email"]
	provider := user.Provider()

	if name == "" || email == "" || provider == "" {
		return nil, sdkerrors.NewInvalidArgumentError("missing user data: [name|email|provider]")
	}
	userID := sdktypes.NewUserIDFromUserData(provider, email, name)

	return &scheme.User{
		UserID:   userID.UUIDValue(),
		Provider: provider,
		Email:    email,
		Name:     name,
	}, nil
}

// extract entity UUID and Type
func entityOwnershipWithIDAndType(entity any) scheme.Ownership {
	switch entity := entity.(type) {
	case scheme.Project:
		return scheme.Ownership{EntityID: entity.ProjectID, EntityType: "Project"}
	case *scheme.Project:
		return scheme.Ownership{EntityID: entity.ProjectID, EntityType: "Project"}
	case scheme.Build:
		return scheme.Ownership{EntityID: entity.BuildID, EntityType: "Build"}
	case *scheme.Build:
		return scheme.Ownership{EntityID: entity.BuildID, EntityType: "Build"}
	case scheme.Deployment:
		return scheme.Ownership{EntityID: entity.DeploymentID, EntityType: "Deployment"}
	case *scheme.Deployment:
		return scheme.Ownership{EntityID: entity.DeploymentID, EntityType: "Deployment"}
	case scheme.Env:
		return scheme.Ownership{EntityID: entity.EnvID, EntityType: "Env"}
	case *scheme.Env:
		return scheme.Ownership{EntityID: entity.EnvID, EntityType: "Env"}
	case scheme.Connection:
		return scheme.Ownership{EntityID: entity.ConnectionID, EntityType: "Connection"}
	case *scheme.Connection:
		return scheme.Ownership{EntityID: entity.ConnectionID, EntityType: "Connection"}
	case scheme.Session:
		return scheme.Ownership{EntityID: entity.SessionID, EntityType: "Session"}
	case *scheme.Session:
		return scheme.Ownership{EntityID: entity.SessionID, EntityType: "Session"}
	case scheme.Event:
		return scheme.Ownership{EntityID: entity.EventID, EntityType: "Event"}
	case *scheme.Event:
		return scheme.Ownership{EntityID: entity.EventID, EntityType: "Event"}
	case scheme.Trigger:
		return scheme.Ownership{EntityID: entity.TriggerID, EntityType: "Trigger"}
	case *scheme.Trigger:
		return scheme.Ownership{EntityID: entity.TriggerID, EntityType: "Trigger"}
	case scheme.Var:
		return scheme.Ownership{EntityID: entity.VarID, EntityType: "Var"}
	case *scheme.Var:
		return scheme.Ownership{EntityID: entity.VarID, EntityType: "Var"}
	}
	return scheme.Ownership{}
}

func prepareOwnershipForEntities1(user *scheme.User, entities ...any) []scheme.Ownership {
	var oo []scheme.Ownership
	for _, entity := range entities {
		if o := entityOwnershipWithIDAndType(entity); o.EntityType != "" {
			o.UserID = user.UserID
			oo = append(oo, o)
		}
	}
	return oo
}

func prepareOwnershipForEntities(ctx context.Context, entities ...any) (*scheme.User, []scheme.Ownership, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}
	oo := prepareOwnershipForEntities1(user, entities...)
	return user, oo, nil
}

func saveOwnershipForEntities(db *gorm.DB, u *scheme.User, oo ...scheme.Ownership) error {
	for _, o := range oo { // sanity. ensure same user
		o.UserID = u.UserID
	}

	// NOTE: should be transactional
	if err := db.Where(u).FirstOrCreate(u).Error; err != nil { // ensure user exists
		return err
	}
	return db.Create(oo).Error // and create ownerships
}

func createEntityWithOwnership[T any](ctx context.Context, db *gorm.DB, model *T) error {
	user, ownerships, err := prepareOwnershipForEntities(ctx, model)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		tx = tx.WithContext(ctx)
		if err := tx.Create(model).Error; err != nil {
			return err
		}

		return saveOwnershipForEntities(tx, user, ownerships...)
	})
}

func (gdb *gormdb) isUserEntity1(user *scheme.User, ids ...sdktypes.UUID) error {
	gdb.z.Debug("isUserEntity", zap.Any("entityIDs", ids), zap.Any("user", user))
	var count int64
	if err := gdb.db.Model(&scheme.Ownership{}).Where("entity_id IN ? AND user_id = ?", ids, user.UserID).Limit(1).Count(&count).Error; err != nil {
		return err
	}
	if count < int64(len(ids)) {
		return sdkerrors.ErrUnauthorized
	}
	// FIXME: could/should we distinguish between not found and unauthorized? Could be useful in updates
	// Is there other way then adding more queries?
	return nil
}

func (gdb *gormdb) isUserEntity(ctx context.Context, ids ...sdktypes.UUID) error {
	user, _ := userFromContext(ctx) // REVIEW: OK to ignore error
	return gdb.isUserEntity1(user, ids...)
}

// add context and join with user ownership on entity
func joinUserEntity(ctx context.Context, db *gorm.DB, entity string, userID sdktypes.UUID) *gorm.DB {
	// REVIEW: the simplest possible way. We could also use generics, TableName, interface to find ID column,  etc..
	tableName := entity + string("s")
	joinExpr := fmt.Sprintf("JOIN ownerships ON ownerships.entity_id = %s.%s_id", tableName, entity)
	return db.WithContext(ctx).
		Table(tableName).Joins(joinExpr).Where("ownerships.user_id = ?", userID)
}

// gorm user+entity scope
// func withUserEntity(ctx context.Context, entity string) func(db *gorm.DB) *gorm.DB {
// 	return func(db *gorm.DB) *gorm.DB {
// 		user, _ := userFromContext(ctx) // FIXME: ignore error?
// 		return joinUserEntity(ctx, db, entity, user.UserID)
// 	}
// }

// gormdb user+entity scoped godm db + logging
func (gdb *gormdb) withUserEntity(ctx context.Context, entity string) *gorm.DB {
	user, _ := userFromContext(ctx) // FIXME: ignore error?
	gdb.z.Debug("withUser", zap.String("entity", entity), zap.Any("user", user))
	return joinUserEntity(ctx, gdb.db, entity, user.UserID)
}

// FIXME:
// 1. ID
// - maybe we could add google/github users earlier? on login?
// - compute userID on user creation and add to proto?
// - could we store emails? (data protection? CCPA/GDPR?)
// - maybe we could rely on providers IDs? we have such for google/github and need one from descope
// - add uniq index for users name+email+provider? or userID is enough which is 1-to-1 to all those 3
// 2. Delete
// - need to check ownership foreign keys for fully deleted objects
// - cleaning ownership table
// - deleting objects
// 3.
// - rmove varID from var and use ScopeID as userID
// - remove line from dbgorm files
// - create
//   - build need to add projectID
//   - connection check with projectID, even if optional
//   - deployment, build should be non-optinal. Need to check vs. buildID or EnvID
//   - env with ProjectID
//   - event if connectionID is present
//   - TRIGGER PROJECTid

// if we are in production do not allow save events
