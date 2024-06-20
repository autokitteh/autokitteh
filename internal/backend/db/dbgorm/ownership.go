package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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
	case sdktypes.Project:
		return scheme.Ownership{EntityID: entity.ID().UUIDValue(), EntityType: "Project"}
	case sdktypes.Build:
		return scheme.Ownership{EntityID: entity.ID().UUIDValue(), EntityType: "Build"}
	case sdktypes.Deployment:
		return scheme.Ownership{EntityID: entity.ID().UUIDValue(), EntityType: "Deployment"}
	case sdktypes.Env:
		return scheme.Ownership{EntityID: entity.ID().UUIDValue(), EntityType: "Env"}
	case sdktypes.Connection:
		return scheme.Ownership{EntityID: entity.ID().UUIDValue(), EntityType: "Connection"}
	case sdktypes.Session:
		return scheme.Ownership{EntityID: entity.ID().UUIDValue(), EntityType: "Session"}
	case sdktypes.Event:
		return scheme.Ownership{EntityID: entity.ID().UUIDValue(), EntityType: "Event"}
	case sdktypes.Trigger:
		return scheme.Ownership{EntityID: entity.ID().UUIDValue(), EntityType: "Trigger"}
	// FIXME: there is no ID method for VAR

	// -------------------------------------------------------------------------------------------
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

func prepareOwnershipForEntities(ctx context.Context, entities ...any) (*scheme.User, []scheme.Ownership, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	var oo []scheme.Ownership
	for _, entity := range entities {
		if o := entityOwnershipWithIDAndType(entity); o.EntityType != "" {
			o.UserID = user.UserID
			oo = append(oo, o)
		}
	}
	return user, oo, nil
}

func saveOwnershipForEntities(ctx context.Context, db *gorm.DB, u *scheme.User, oo ...scheme.Ownership) error {
	for _, o := range oo { // sanity. ensure same user
		o.UserID = u.UserID
	}

	// FIXME: add transaction - will require reqursive one
	db = db.WithContext(ctx)
	if err := db.Where(u).FirstOrCreate(u).Error; err != nil {
		return err
	}

	return db.Create(oo).Error
}

func createEntityWithOwnership[T any](ctx context.Context, gdb *gormdb, model *T, createFunc func(*T) error) error {
	user, ownerships, err := prepareOwnershipForEntities(ctx, model)
	if err != nil {
		return err
	}
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := createFunc(model); err != nil {
			return err
		}

		return saveOwnershipForEntities(ctx, tx.gormdb.db, user, ownerships...)
	})
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
// - add uniq index for users name+email+provider? or userID is enough which is 1-to-1 to all those 3
