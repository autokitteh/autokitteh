package dbgorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type (
	idFieldNamer   interface{ IDFieldName() string }
	getProjectIDer interface{ GetProjectID() sdktypes.ProjectID }
	uuidValuer     interface{ UUIDValue() uuid.UUID }
)

func (gdb *gormdb) getRecordOwner(ctx context.Context, m idFieldNamer, id uuidValuer) (sdktypes.OwnerID, error) {
	var o scheme.Owned

	err := gdb.db.WithContext(ctx).
		Model(m).
		Where(fmt.Sprintf("%s = ?", m.IDFieldName()), id.UUIDValue()).
		Select("owner_user_id", "owner_org_id").
		First(&o).
		Error
	if err != nil {
		return sdktypes.InvalidOwnerID, translateError(err)
	}

	return o.GetOwnerID(), nil
}

func (gdb *gormdb) getRecordProjectOwner(
	ctx context.Context,
	m interface {
		idFieldNamer
		getProjectIDer
	},
	id uuidValuer,
) (sdktypes.OwnerID, error) {
	var p scheme.BelongsToProject

	err := gdb.db.WithContext(ctx).
		Model(m).
		Where(fmt.Sprintf("%s = ?", m.IDFieldName()), id.UUIDValue()).
		Preload("Project").
		Select("project_id").
		First(&p).
		Error
	if err != nil {
		return sdktypes.InvalidOwnerID, translateError(err)
	}

	return p.Project.GetOwnerID(), nil
}

func (gdb *gormdb) GetOwner(ctx context.Context, id sdktypes.ID) (sdktypes.OwnerID, error) {
	switch id.Kind() {
	case sdktypes.BuildIDKind:
		return gdb.getRecordOwner(ctx, scheme.Build{}, id)
	case sdktypes.SessionIDKind:
		return gdb.getRecordOwner(ctx, scheme.Session{}, id)
	case sdktypes.ProjectIDKind:
		return gdb.getRecordOwner(ctx, scheme.Project{}, id)
	case sdktypes.ConnectionIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Connection{}, id)
	case sdktypes.TriggerIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Trigger{}, id)
	case sdktypes.DeploymentIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Deployment{}, id)
	case sdktypes.EventIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Event{}, id)
	case sdktypes.UserIDKind:
		return sdktypes.NewOwnerID(sdktypes.FromID[sdktypes.UserID](id)), nil
	case sdktypes.OrgIDKind:
		return sdktypes.NewOwnerID(sdktypes.FromID[sdktypes.OrgID](id)), nil
	default:
		return sdktypes.InvalidOwnerID, sdkerrors.NewInvalidArgumentError("unhandled id kind")
	}
}

func (gdb *gormdb) GetProjectID(ctx context.Context, id sdktypes.ID) (sdktypes.ProjectID, error) {
	var m idFieldNamer

	// Only records that project is mandatory in them.
	switch id.Kind() {
	case sdktypes.ConnectionIDKind:
		m = scheme.Connection{}
	case sdktypes.TriggerIDKind:
		m = scheme.Trigger{}
	case sdktypes.DeploymentIDKind:
		m = scheme.Deployment{}
	case sdktypes.EventIDKind:
		m = scheme.Event{}
	default:
		return sdktypes.InvalidProjectID, sdkerrors.NewInvalidArgumentError("unhandled id kind")
	}

	var p scheme.BelongsToProject

	err := gdb.db.WithContext(ctx).
		Model(m).
		Where(fmt.Sprintf("%s = ?", m.IDFieldName()), id.UUIDValue()).
		Select("project_id").
		First(&p).
		Error
	if err != nil {
		return sdktypes.InvalidProjectID, translateError(err)
	}

	return p.GetProjectID(), nil
}
