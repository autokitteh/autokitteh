package dbgorm

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const maxValueSize = 64 * 1024

func (db *gormdb) SetValue(ctx context.Context, envID sdktypes.EnvID, key string, v sdktypes.Value) error {
	if !envID.IsValid() {
		return sdkerrors.NewInvalidArgumentError("invalid env_id")
	}

	if v.ProtoSize() > maxValueSize {
		return sdkerrors.NewInvalidArgumentError("value too large > %d bytes", maxValueSize)
	}

	uid, err := userIDFromContext(ctx)
	if err != nil {
		return err
	}

	return translateError(db.transaction(ctx, func(tx *tx) error {
		db := tx.db

		oo, err := tx.owner.EnsureUserAccessToEntitiesWithOwnership(tx.ctx, db, uid, envID.UUIDValue())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return gorm.ErrForeignKeyViolated
			}
			return err
		}
		if len(oo) != 1 {
			return gorm.ErrForeignKeyViolated
		}

		bs, err := proto.Marshal(v.ToProto())
		if err != nil {
			return err
		}

		return tx.db.Save(&scheme.Value{EnvID: envID.UUIDValue(), Key: key, Value: bs}).Error
	}))
}

func (db *gormdb) GetValue(ctx context.Context, envID sdktypes.EnvID, key string) (sdktypes.Value, error) {
	r, err := getOne[scheme.Value](db.withUserEnvs(ctx), "env_id = ? AND key = ?", envID.String(), key)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	var pb sdktypes.ValuePB

	if err := proto.Unmarshal(r.Value, &pb); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.ValueFromProto(&pb)
}

func (db *gormdb) ListValues(ctx context.Context, envID sdktypes.EnvID) (map[string]sdktypes.Value, error) {
	var rs []*scheme.Value
	if err := db.withUserEnvs(ctx).WithContext(ctx).Where("env_id = ?", envID.UUIDValue()).Find(&rs).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.ListToMapError(rs, func(r *scheme.Value) (string, sdktypes.Value, error) {
		var pb sdktypes.ValuePB

		if err := proto.Unmarshal(r.Value, &pb); err != nil {
			return "", sdktypes.InvalidValue, err
		}

		v, err := sdktypes.ValueFromProto(&pb)
		if err != nil {
			return "", sdktypes.InvalidValue, err
		}

		return r.Key, v, err
	})
}
