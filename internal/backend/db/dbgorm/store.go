package dbgorm

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const maxValueSize = 64 * 1024

func (db *gormdb) SetStoreValue(ctx context.Context, pid sdktypes.ProjectID, key string, v sdktypes.Value) error {
	if !pid.IsValid() {
		return sdkerrors.NewInvalidArgumentError("invalid project id")
	}

	q := db.writer.WithContext(ctx)

	if !v.IsValid() {
		return translateError(
			q.Where("project_id = ? AND key = ?", pid.UUIDValue(), key).Delete(&scheme.StoreValue{}).Error,
		)
	}

	if v.ProtoSize() > maxValueSize {
		return sdkerrors.NewInvalidArgumentError("value too large > %d bytes", maxValueSize)
	}

	bs, err := proto.Marshal(v.ToProto())
	if err != nil {
		return err
	}

	by := authcontext.GetAuthnUserID(ctx).UUIDValue()

	err = q.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "project_id"}, {Name: "key"}},
		DoUpdates: clause.Assignments(map[string]any{
			"value":      bs,
			"updated_by": by,
			"updated_at": kittehs.Now(),
			// DO NOT reset "published" flag on update.
		}),
	}).Create(&scheme.StoreValue{
		Base:      based(ctx),
		ProjectID: pid.UUIDValue(),
		Key:       key,
		Value:     bs,
		UpdatedBy: by,
		UpdatedAt: kittehs.Now(),
	}).Error

	return translateError(err)
}

func (db *gormdb) GetStoreValue(ctx context.Context, pid sdktypes.ProjectID, key string) (sdktypes.Value, error) {
	kvs, err := db.ListStoreValues((ctx), pid, []string{key}, true)
	if err != nil {
		return sdktypes.InvalidValue, translateError(err)
	}
	if len(kvs) == 0 {
		return sdktypes.Nothing, nil
	}

	v, ok := kvs[key]
	if !ok {
		return sdktypes.InvalidValue, errors.New("internal error: key not found in kvs")
	}

	return v, nil
}

func (db *gormdb) PublishStoreValue(ctx context.Context, pid sdktypes.ProjectID, key string) error {
	if !pid.IsValid() {
		return sdkerrors.NewInvalidArgumentError("invalid project id")
	}

	q := db.writer.WithContext(ctx).Model(&scheme.StoreValue{}).Where("project_id = ? AND key = ?", pid.UUIDValue(), key).Update("published", true)
	if err := q.Error; err != nil {
		return translateError(err)
	}

	if q.RowsAffected == 0 {
		return sdkerrors.ErrNotFound
	}

	return nil
}

func (db *gormdb) IsStoreValuePublished(ctx context.Context, pid sdktypes.ProjectID, key string) (bool, error) {
	var sv scheme.StoreValue
	err := db.reader.WithContext(ctx).Select("published").Where("project_id = ? AND key = ?", pid.UUIDValue(), key).First(&sv).Error
	if err != nil {
		return false, translateError(err)
	}

	return sv.Published, nil
}

func (db *gormdb) ListStoreValues(ctx context.Context, pid sdktypes.ProjectID, keys []string, getValues bool) (map[string]sdktypes.Value, error) {
	var rs []*scheme.StoreValue
	q := db.reader.WithContext(ctx).Where("project_id = ?", pid.UUIDValue())

	if len(keys) > 0 {
		q = q.Where("key IN (?)", keys)
	}

	if !getValues {
		q = q.Select("key")
	}

	if err := q.Find(&rs).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.ListToMapError(rs, func(r *scheme.StoreValue) (string, sdktypes.Value, error) {
		if !getValues {
			return r.Key, sdktypes.InvalidValue, nil
		}

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
