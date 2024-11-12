package dbgorm

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const maxValueSize = 64 * 1024

func (db *gormdb) SetValue(ctx context.Context, pid sdktypes.ProjectID, key string, v sdktypes.Value) error {
	if !pid.IsValid() {
		return sdkerrors.NewInvalidArgumentError("invalid project id")
	}

	if v.ProtoSize() > maxValueSize {
		return sdkerrors.NewInvalidArgumentError("value too large > %d bytes", maxValueSize)
	}

	uid := authcontext.GetAuthnUser(ctx).ID().String()

	return translateError(db.transaction(ctx, func(tx *tx) error {
		db := tx.db

		oo, err := tx.owner.EnsureUserAccessToEntitiesWithOwnership(tx.ctx, db, uid, pid.UUIDValue())
		if err != nil {
			return err
		}
		if len(oo) != 1 {
			return gorm.ErrForeignKeyViolated
		}

		bs, err := proto.Marshal(v.ToProto())
		if err != nil {
			return err
		}

		return tx.db.Save(&scheme.Value{ProjectID: pid.UUIDValue(), Key: key, Value: bs, UpdatedAt: time.Now().UTC()}).Error
	}))
}

func (db *gormdb) GetValue(ctx context.Context, pid sdktypes.ProjectID, key string) (sdktypes.Value, error) {
	r, err := getOne[scheme.Value](db.withUserProjects(ctx), "project_id = ? AND key = ?", pid.UUIDValue(), key)
	if err != nil {
		return sdktypes.InvalidValue, translateError(err)
	}

	var pb sdktypes.ValuePB

	if err := proto.Unmarshal(r.Value, &pb); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.ValueFromProto(&pb)
}

func (db *gormdb) ListValues(ctx context.Context, pid sdktypes.ProjectID) (map[string]sdktypes.Value, error) {
	var rs []*scheme.Value
	if err := db.withUserProjects(ctx).Where("project_id = ?", pid.UUIDValue()).Find(&rs).Error; err != nil {
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
