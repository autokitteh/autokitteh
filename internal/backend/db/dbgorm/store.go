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

	return translateError(
		q.Save(&scheme.StoreValue{
			// This does not take into account if value was also previously set,
			// so `created_by` and `updated_by` would always point to the last user
			// that updated the user.
			// Getting `created_by` correctly would neccessitate another query, not
			// worth it currently.
			Base:      based(ctx),
			ProjectID: pid.UUIDValue(),
			Key:       key,
			Value:     bs,
		}).Error,
	)
}

func (db *gormdb) GetStoreValue(ctx context.Context, pid sdktypes.ProjectID, key string) (sdktypes.Value, error) {
	r, err := getOne[scheme.StoreValue](db.reader.WithContext(ctx), "project_id = ? AND key = ?", pid.UUIDValue(), key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sdktypes.Nothing, nil
	}

	if err != nil {
		return sdktypes.InvalidValue, translateError(err)
	}

	var pb sdktypes.ValuePB

	if err := proto.Unmarshal(r.Value, &pb); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.ValueFromProto(&pb)
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
