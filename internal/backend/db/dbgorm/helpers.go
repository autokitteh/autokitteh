package dbgorm

import (
	"context"
	"fmt"
	"maps"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func uuidPtrOrNil(o interface{ UUIDValue() uuid.UUID }) *uuid.UUID {
	v := o.UUIDValue()

	if v == uuid.Nil {
		return nil
	}

	return &v
}

func withProjectID(q *gorm.DB, field string, pid sdktypes.ProjectID) *gorm.DB {
	if !pid.IsValid() {
		return q
	}

	if field != "" {
		field += "."
	}

	return q.Where(field+"project_id = ?", pid.UUIDValue())
}

func withProjectOrgID(q *gorm.DB, oid sdktypes.OrgID, table string) *gorm.DB {
	if !oid.IsValid() {
		return q
	}

	return q.Joins(fmt.Sprintf("INNER JOIN projects ON %s.project_id = projects.project_id AND projects.org_id = ?", table), oid.UUIDValue()).Distinct()
}

func based(ctx context.Context) scheme.Base {
	now := kittehs.Now().UTC()

	uid := authcontext.GetAuthnUserID(ctx).UUIDValue()

	return scheme.Base{
		CreatedBy: uid,
		CreatedAt: now,
	}
}

func updatedBaseColumns(ctx context.Context) map[string]any {
	m := map[string]any{"updated_at": kittehs.Now().UTC()}

	if uid := authcontext.GetAuthnUserID(ctx); uid.IsValid() {
		m["updated_by"] = authcontext.GetAuthnUserID(ctx).UUIDValue()
	}

	return m
}

func updateBaseColumns(ctx context.Context, m map[string]any) {
	maps.Copy(m, updatedBaseColumns(ctx))
}

type updateableMessage[M proto.Message] interface {
	ToProto() M
	IsMutableField(string) bool
	Mutables() []string
}

func updatedFields[M proto.Message](
	ctx context.Context,
	m updateableMessage[M],
	fm *sdktypes.FieldMask,
) (data map[string]any, err error) {
	if fm == nil || len(fm.Paths) == 0 {
		// If no field mask is provided, update all mutable fields.
		fm = &sdktypes.FieldMask{Paths: m.Mutables()}
	}

	// Normalize a copy of the field mask so not to mutate the original.
	fm = &sdktypes.FieldMask{Paths: fm.Paths}
	fm.Normalize()

	data, err = kittehs.ProtoToMap(m.ToProto(), fm)
	if err != nil {
		return nil, sdkerrors.NewInvalidArgumentError("invalid field mask")
	}

	for k := range data {
		if !m.IsMutableField(k) {
			return nil, sdkerrors.NewInvalidArgumentError("field %q is not mutable", k)
		}
	}

	updateBaseColumns(ctx, data)

	return data, nil
}
