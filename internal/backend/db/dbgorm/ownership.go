package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gormdb *gormdb) AddOwnership(ctx context.Context, entities ...any) error {
	userID, err := gormdb.EnsureOwnershipUser(ctx)
	if err != nil {
		return err
	}
	ownership := scheme.Ownership{
		UserID: userID.UUIDValue(),
	}

	for _, e := range entities {
		var uuid *sdktypes.UUID
		var entityType string

		switch entity := e.(type) {
		case sdktypes.Project:
			uuid = entity.ID().Value()
			entityType = "Project"
		case sdktypes.Build:
			uuid = entity.ID().Value()
			entityType = "Build"
		case sdktypes.Deployment:
			uuid = entity.ID().Value()
			entityType = "Deployment"
		case sdktypes.Env:
			uuid = entity.ID().Value()
			entityType = "Env"
		case sdktypes.Connection:
			uuid = entity.ID().Value()
			entityType = "Connection"
		case sdktypes.Session:
			uuid = entity.ID().Value()
			entityType = "Session"
		case sdktypes.Event:
			uuid = entity.ID().Value()
			entityType = "Event"
		case sdktypes.Trigger:
			uuid = entity.ID().Value()
			entityType = "Trigger"
		case sdktypes.Var:
			uuid = entity.ScopeID().Value()
			entityType = "Var"
		}

		if uuid != nil {
			ownership.EntityID = *uuid
			ownership.EntityType = entityType
			if err = gormdb.db.WithContext(ctx).Where(ownership).FirstOrCreate(&ownership).Error; err != nil {
				return translateError(err)
			}
		}
	}
	return nil
}

func (db *gormdb) EnsureOwnershipUser(ctx context.Context) (sdktypes.UserID, error) {
	user := authcontext.GetAuthnUser(ctx)
	if !user.IsValid() {
		user = sdktypes.DefaultUser
	}

	data := user.Data()
	name := data["name"]
	email := data["email"]
	provider := user.Provider()

	if name == "" || email == "" || provider == "" {
		return sdktypes.InvalidUserID, sdkerrors.NewInvalidArgumentError("missing user data: [name|email|provider]")
	}
	userID := sdktypes.NewUserIDFromUserData(provider, email, name)

	u := scheme.User{
		UserID:   userID.UUIDValue(),
		Provider: provider,
		Email:    email,
		Name:     name,
	}
	err := db.db.WithContext(ctx).Where(u).FirstOrCreate(&u).Error

	return userID, translateError(err)
}

// FIXME:
// - maybe we could add google/github users earlier? on login?
// - compute userID on user creation and add to proto?
// - could we store emails? (data protection? CCPA/GDPR?)
// - maybe we could rely on providers IDs? we have such for google/github and need one from descope
// - ensureOwnershipUser. Here under transaction (nested transaction for projects, deployments, events) or in scheme
// - we could add Addownership on db layer or on service layer. Meanwhile added on services layer
// - we could add a auth wrapper on top of regular func or just insert logic inside. Auth wrapper will require nested transcation
// - need to check ownership foreign keys for fully deleted objects
// - deleting objects
