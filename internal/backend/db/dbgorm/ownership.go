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

func (gdb *gormdb) createEntityWithOwnership(
	ctx context.Context, create func(tx *gorm.DB, user *scheme.User) error, model any, allowedOnID ...*sdktypes.UUID,
) error {
	user, ownerships, err := prepareOwnershipForEntities(ctx, model)
	if err != nil {
		return err
	}

	var idsToVerifyOwnership []sdktypes.UUID
	if len(allowedOnID) != 0 {
		uniqueIDs := make(map[sdktypes.UUID]struct{}, len(allowedOnID))
		for _, id := range allowedOnID {
			if id != nil {
				uniqueIDs[*id] = struct{}{}
			}
		}
		for id := range uniqueIDs {
			idsToVerifyOwnership = append(idsToVerifyOwnership, id)
		}
	}

	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isUserEntity1(user, idsToVerifyOwnership...); err != nil {
			return err
		}
		if err := create(tx.db, user); err != nil { // create
			return err
		}
		return saveOwnershipForEntities(tx.db, user, ownerships...) // ensure user and add ownerships
	})
}

func (gdb *gormdb) isUserEntity1(user *scheme.User, ids ...sdktypes.UUID) error {
	gdb.z.Debug("isUserEntity", zap.Any("entityIDs", ids), zap.Any("user", user))
	if len(ids) == 0 {
		return nil
	}

	var oo []scheme.Ownership
	if err := gdb.db.Model(&scheme.Ownership{}).Where("entity_id IN ?", ids).Select("user_id").Find(&oo).Error; err != nil {
		return err
	}

	if len(oo) < len(ids) {
		return gorm.ErrRecordNotFound
	}

	for _, o := range oo {
		if o.UserID != user.UserID {
			return sdkerrors.ErrUnauthorized
		}
	}
	return nil
}

func (gdb *gormdb) isUserEntity(ctx context.Context, ids ...sdktypes.UUID) error {
	user, _ := userFromContext(ctx) // REVIEW: OK to ignore error
	return gdb.isUserEntity1(user, ids...)
}

// REVIEW: this is probably the simplest possible way (e.g. with entity as string).
// We could also use generics, TableName, interface to find ID column,  etc..

// join with user ownership on entity
func joinUserEntity(db *gorm.DB, entity string, userID sdktypes.UUID) *gorm.DB {
	tableName := entity + string("s")
	joinExpr := fmt.Sprintf("JOIN ownerships ON ownerships.entity_id = %s.%s_id", tableName, entity)
	return db.Table(tableName).Joins(joinExpr).Where("ownerships.user_id = ?", userID)
}

// gorm user+entity scope
func withUserEntity(entity string, userID sdktypes.UUID) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return joinUserEntity(db, entity, userID)
	}
}

// gormdb user+entity scoped godm db + logging
func (gdb *gormdb) withUserEntity(ctx context.Context, entity string) *gorm.DB {
	user, _ := userFromContext(ctx) // NOTE: ignore possible error
	gdb.z.Debug("withUser", zap.String("entity", entity), zap.Any("user", user))
	db := gdb.db.WithContext(ctx)
	return joinUserEntity(db, entity, user.UserID)
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
// - cleaning ownership table? deleting objects
// 3.
//   if we are in production do not allow save events via cmd?
// 4. tests
// - list/delete
