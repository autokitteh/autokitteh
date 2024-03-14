package dbgorm

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/db/dbgorm/scheme"
)

func (db *gormdb) CreateUser(ctx context.Context, user sdktypes.User) error {
	if !user.ID().IsValid() {
		db.z.DPanic("no user id supplied")
		return errors.New("user id missing")
	}

	r := scheme.User{
		UserID: user.ID().String(),
		Name:   user.Name().String(),
	}

	if err := db.db.WithContext(ctx).Create(&r).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) GetUserByID(ctx context.Context, uid sdktypes.UserID) (sdktypes.User, error) {
	// passing scheme.ParserUser makes this function highly unreadable
	return get(db.db, ctx, scheme.ParseUser, "user_id = ?", uid.String())
}

func (db *gormdb) GetUserByName(ctx context.Context, name sdktypes.Symbol) (sdktypes.User, error) {
	return get(db.db, ctx, scheme.ParseUser, "name = ?", name.String())
}

func (db *gormdb) GetUserByExternalID(ctx context.Context, externalID string) (sdktypes.User, error) {
	var userExternalID *scheme.UserExternalIdentitiy
	result := db.db.WithContext(ctx).Preload("User").Where("external_id = ?", externalID).First(&userExternalID)
	if result.Error != nil {
		return sdktypes.InvalidUser, translateError(result.Error)
	}

	if result.RowsAffected == 0 {
		return sdktypes.InvalidUser, sdkerrors.ErrNotFound
	}

	return scheme.ParseUser(userExternalID.User)
}

func (db *gormdb) AddExternalIDToUser(ctx context.Context, uid sdktypes.UserID, externalID, idType, idEmail string) error {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	eID := scheme.UserExternalIdentitiy{
		UserExternalIdentitiyID: newUUID.String(),
		ExternalID:              externalID,
		UserID:                  uid.String(),
		IdentityType:            idType,
		Email:                   idEmail,
	}

	if err := db.db.WithContext(ctx).Create(&eID).Error; err != nil {
		return translateError(err)
	}

	return nil
}
