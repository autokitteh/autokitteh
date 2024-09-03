package dbgorm

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	akCtx "go.autokitteh.dev/autokitteh/internal/backend/context"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) GetOwnership(ctx context.Context, entityID sdktypes.UUID) (uid string, err error) {
	var o scheme.Ownership
	if err = gdb.db.WithContext(ctx).Where("entity_id = ?", entityID).Select("user_id").First(&o).Error; err != nil {
		gdb.z.Error("get ownership", zap.String("id", entityID.String()), zap.Error(err))
		return
	}
	return o.UserID, nil
}

func userIDFromContext(ctx context.Context) (string, error) {
	if uid := authcontext.GetAuthnUserID(ctx); uid != "" {
		return uid, nil
	}

	user := authcontext.GetAuthnUser(ctx)
	if !user.IsValid() {
		return "", sdkerrors.NewInvalidArgumentError("unknown user")
	}

	data := user.Data()
	name := data["name"]
	email := data["email"]
	provider := user.Provider()

	if name == "" || email == "" || provider == "" {
		return "", sdkerrors.NewInvalidArgumentError("unknown user")
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
	var ownerships []scheme.Ownership
	for _, entity := range entities {
		if o := entityOwnershipWithIDAndType(entity); o.EntityType != "" {
			o.UserID = uid
			ownerships = append(ownerships, o)
		}
	}
	return ownerships
}

func prepareOwnershipForEntities(ctx context.Context, entities ...any) (string, []scheme.Ownership, error) {
	uid, err := userIDFromContext(ctx)
	if err != nil {
		return "", nil, err
	}
	oo := prepareOwnershipForEntities1(uid, entities...)
	return uid, oo, nil
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
	uid, ownerships, err := prepareOwnershipForEntities(ctx, model)
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
		if err := tx.isUserEntity(*tx.ctx, uid, idsToVerifyOwnership...); err != nil {
			return err
		}
		if err := create(tx.db, uid); err != nil { // create
			return err
		}
		return saveOwnershipForEntities(tx.db, uid, ownerships...) // ensure user and add ownerships
	})
}

func getOwnershipsForEntities(db *gorm.DB, ids ...sdktypes.UUID) ([]scheme.Ownership, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var ownerships []scheme.Ownership
	if err := db.Model(&scheme.Ownership{}).Where("entity_id IN ?", ids).Find(&ownerships).Error; err != nil {
		return nil, err
	}
	return ownerships, nil
}

func verifyOwnerships(uid string, ownerships []scheme.Ownership) error {
	for _, o := range ownerships {
		if o.UserID != uid {
			return sdkerrors.ErrUnauthorized
		}
	}
	return nil
}

// fetches ownerships for given ids and verifies that user have access to all of them
func ensureUserAccessToEntitiesWithOwnerships(db *gorm.DB, uid string, ids ...sdktypes.UUID) ([]scheme.Ownership, error) {
	ownerships, err := getOwnershipsForEntities(db, ids...)
	if err != nil {
		return nil, err
	}

	if len(ownerships) < len(ids) { // should be equal
		return nil, gorm.ErrRecordNotFound
	}
	return ownerships, verifyOwnerships(uid, ownerships)
}

func (gdb *gormdb) isUserEntity(ctx context.Context, uid string, ids ...sdktypes.UUID) error {
	if akCtx.RequestOrginator(ctx) == akCtx.User { // enforce only on user-orginated requests
		return gdb.owner.EnsureUserAccessToEntities(ctx, gdb.db, uid, ids...)
	}
	return nil
}

func (gdb *gormdb) isCtxUserEntity(ctx context.Context, ids ...sdktypes.UUID) error {
	uid, _ := userIDFromContext(ctx)
	return gdb.isUserEntity(ctx, uid, ids...)
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
func withUserEntity(ctx context.Context, gdb *gormdb, entity string, uid string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return gdb.owner.JoinUserEntity(ctx, db, entity, uid)
	}
}

// gormdb user+entity scoped godm db + logging
func (gdb *gormdb) withUserEntity(ctx context.Context, entity string) *gorm.DB {
	user, _ := userIDFromContext(ctx) // NOTE: ignore possible error
	return gdb.owner.JoinUserEntity(ctx, gdb.db.WithContext(ctx), entity, user)
}

type OwnershipChecker interface {
	EnsureUserAccessToEntities(ctx context.Context, db *gorm.DB, user string, ids ...sdktypes.UUID) error
	EnsureUserAccessToEntitiesWithOwnership(ctx context.Context, db *gorm.DB, user string, ids ...sdktypes.UUID) ([]scheme.Ownership, error)
	JoinUserEntity(ctx context.Context, db *gorm.DB, entity string, user string) *gorm.DB
}

type UsersOwnershipChecker struct {
	z *zap.Logger
}

func (c *UsersOwnershipChecker) EnsureUserAccessToEntitiesWithOwnership(
	ctx context.Context, db *gorm.DB, uid string, ids ...sdktypes.UUID,
) (ownerships []scheme.Ownership, err error) {
	c.z.Debug("isUserEntity", zap.String("origin", akCtx.RequestOrginator(ctx).String()), zap.Any("entityIDs", ids), zap.Any("uid", uid))
	ownerships, err = ensureUserAccessToEntitiesWithOwnerships(db, uid, ids...)
	if akCtx.RequestOrginator(ctx) != akCtx.User && errors.Is(err, sdkerrors.ErrUnauthorized) {
		err = nil // ignore not authorized, but keep all other (e.g. NotFound) - see var
	}
	return
}

func (c *UsersOwnershipChecker) EnsureUserAccessToEntities(ctx context.Context, db *gorm.DB, uid string, ids ...sdktypes.UUID) error {
	_, err := c.EnsureUserAccessToEntitiesWithOwnership(ctx, db, uid, ids...)
	return err
}

func (c *UsersOwnershipChecker) JoinUserEntity(ctx context.Context, db *gorm.DB, entity string, uid string) *gorm.DB {
	c.z.Debug("withUser", zap.String("origin", akCtx.RequestOrginator(ctx).String()), zap.String("entity", entity), zap.Any("uid", uid))
	if akCtx.RequestOrginator(ctx) == akCtx.User {
		return joinUserEntity(db, entity, uid)
	}
	return db
}

type PermissiveOwnershipChecker struct {
	z *zap.Logger
}

func (c *PermissiveOwnershipChecker) EnsureUserAccessToEntitiesWithOwnership(
	ctx context.Context, db *gorm.DB, uid string, ids ...sdktypes.UUID,
) ([]scheme.Ownership, error) {
	oo, err := ensureUserAccessToEntitiesWithOwnerships(db, uid, ids...)
	if err != nil && errors.Is(err, sdkerrors.ErrUnauthorized) {
		err = nil // ignore not authorized, but keep all other (e.g. NotFound) - see var
	}
	return oo, err
}

func (c *PermissiveOwnershipChecker) EnsureUserAccessToEntities(ctx context.Context, db *gorm.DB, uid string, ids ...sdktypes.UUID) error {
	return nil
}

func (c *PermissiveOwnershipChecker) JoinUserEntity(ctx context.Context, db *gorm.DB, entity string, uid string) *gorm.DB {
	return db
}
