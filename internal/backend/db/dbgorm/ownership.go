package dbgorm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func userIDFromContext(ctx context.Context) (string, error) {
	user := authcontext.GetAuthnUser(ctx)
	if !user.IsValid() {
		user = sdktypes.DefaultUser
	}

	data := user.Data()
	name := data["name"]
	email := data["email"]
	provider := user.Provider()

	if name == "" || email == "" || provider == "" {
		return "", sdkerrors.NewInvalidArgumentError("missing user data: [name|email|provider]")
	}

	return sdktypes.NewUserIDFromUserData(provider, email, name).String(), nil
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

func prepareOwnershipForEntities1(uid string, entities ...any) []scheme.Ownership {
	var oo []scheme.Ownership
	for _, entity := range entities {
		if o := entityOwnershipWithIDAndType(entity); o.EntityType != "" {
			o.UserID = uid
			oo = append(oo, o)
		}
	}
	return oo
}

func prepareOwnershipForEntities(ctx context.Context, entities ...any) (string, []scheme.Ownership, error) {
	user, err := userIDFromContext(ctx)
	if err != nil {
		return "", nil, err
	}
	oo := prepareOwnershipForEntities1(user, entities...)
	return user, oo, nil
}

func saveOwnershipForEntities(db *gorm.DB, uid string, oo ...scheme.Ownership) error {
	for i := range oo { // sanity. ensure same user
		oo[i].UserID = uid
	}

	return db.Create(oo).Error // and create ownerships
}

func (gdb *gormdb) createEntityWithOwnership(
	ctx context.Context, create func(tx *gorm.DB, uid string) error, model any, allowedOnID ...*sdktypes.UUID,
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
		if err := tx.isUserEntity(user, idsToVerifyOwnership...); err != nil {
			return err
		}
		if err := create(tx.db, user); err != nil { // create
			return err
		}
		return saveOwnershipForEntities(tx.db, user, ownerships...) // ensure user and add ownerships
	})
}

func isUserEntity(db *gorm.DB, uid string, ids ...sdktypes.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	var oo []scheme.Ownership
	if err := db.Model(&scheme.Ownership{}).Where("entity_id IN ?", ids).Select("user_id").Find(&oo).Error; err != nil {
		return err
	}

	if len(oo) < len(ids) {
		return gorm.ErrRecordNotFound
	}

	for _, o := range oo {
		if o.UserID != uid {
			return sdkerrors.ErrUnauthorized
		}
	}
	return nil
}

func (gdb *gormdb) isUserEntity(uid string, ids ...sdktypes.UUID) error {
	return gdb.owner.IsUserEntity(gdb.db, uid, ids...)
}

func (gdb *gormdb) isCtxUserEntity(ctx context.Context, ids ...sdktypes.UUID) error {
	uid, _ := userIDFromContext(ctx)
	return gdb.isUserEntity(uid, ids...)
}

// REVIEW: this is probably the simplest possible way (e.g. with entity as string).
// We could also use generics, TableName, interface to find ID column,  etc..

// join with user ownership on entity
func joinUserEntity(db *gorm.DB, entity string, uid string) *gorm.DB {
	tableName := entity + string("s")
	joinExpr := fmt.Sprintf("JOIN ownerships ON ownerships.entity_id = %s.%s_id", tableName, entity)
	return db.Table(tableName).Joins(joinExpr).Where("ownerships.user_id = ?", uid)
}

// gorm user+entity scope
func withUserEntity(gdb *gormdb, entity string, uid string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return gdb.owner.JoinUserEntity(db, entity, uid)
	}
}

// gormdb user+entity scoped godm db + logging
func (gdb *gormdb) withUserEntity(ctx context.Context, entity string) *gorm.DB {
	user, _ := userIDFromContext(ctx) // NOTE: ignore possible error
	return gdb.owner.JoinUserEntity(gdb.db.WithContext(ctx), entity, user)
}

type OwnershipChecker interface {
	IsUserEntity(db *gorm.DB, user string, ids ...sdktypes.UUID) error
	JoinUserEntity(db *gorm.DB, entity string, user string) *gorm.DB
}

type UsersOwnershipChecker struct {
	z *zap.Logger
}

func (c *UsersOwnershipChecker) IsUserEntity(db *gorm.DB, uid string, ids ...sdktypes.UUID) error {
	c.z.Debug("isUserEntity", zap.Any("entityIDs", ids), zap.Any("uid", uid))
	return isUserEntity(db, uid, ids...)
}

func (c *UsersOwnershipChecker) JoinUserEntity(db *gorm.DB, entity string, uid string) *gorm.DB {
	c.z.Debug("withUser", zap.String("entity", entity), zap.Any("uid", uid))
	return joinUserEntity(db, entity, uid)
}

type PermissiveOwnershipChecker struct {
	z *zap.Logger
}

func (c *PermissiveOwnershipChecker) IsUserEntity(db *gorm.DB, uid string, ids ...sdktypes.UUID) error {
	return nil
}

func (c *PermissiveOwnershipChecker) JoinUserEntity(db *gorm.DB, entity string, uid string) *gorm.DB {
	return db
}
