package dbgorm

import (
	"context"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type (
	idFieldNamer interface{ IDFieldName() string }
	uuidValuer   interface{ UUIDValue() uuid.UUID }
)

func (gdb *gormdb) getProjectOrg(ctx context.Context, id uuidValuer) (sdktypes.OrgID, error) {
	var p scheme.Project

	err := gdb.rdb.WithContext(ctx).
		Where("project_id = ?", id.UUIDValue()).
		Select("org_id").
		First(&p).
		Error
	if err != nil {
		return sdktypes.InvalidOrgID, translateError(err)
	}

	return sdktypes.NewIDFromUUID[sdktypes.OrgID](p.OrgID), nil
}

func (gdb *gormdb) getRecordProjectOwner(
	ctx context.Context,
	m idFieldNamer,
	id uuidValuer,
) (sdktypes.OrgID, error) {
	var p struct {
		ProjectID uuid.UUID      `gorm:"column:project_id"`
		Project   scheme.Project `gorm:"foreignKey:project_id"`
	}

	err := gdb.rdb.WithContext(ctx).
		Model(m).
		Where(m.IDFieldName()+" = ?", id.UUIDValue()).
		Preload("Project").
		Select("project_id").
		First(&p).
		Error
	if err != nil {
		return sdktypes.InvalidOrgID, translateError(err)
	}

	return sdktypes.NewIDFromUUID[sdktypes.OrgID](p.Project.OrgID), nil
}

func (gdb *gormdb) GetOrgIDOf(ctx context.Context, id sdktypes.ID) (sdktypes.OrgID, error) {
	switch id.Kind() {
	case sdktypes.OrgIDKind:
		return sdktypes.FromID[sdktypes.OrgID](id), nil
	case sdktypes.ProjectIDKind:
		return gdb.getProjectOrg(ctx, id)
	case sdktypes.BuildIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Build{}, id)
	case sdktypes.SessionIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Session{}, id)
	case sdktypes.ConnectionIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Connection{}, id)
	case sdktypes.TriggerIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Trigger{}, id)
	case sdktypes.DeploymentIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Deployment{}, id)
	case sdktypes.EventIDKind:
		return gdb.getRecordProjectOwner(ctx, scheme.Event{}, id)
	case sdktypes.IntegrationIDKind, sdktypes.UserIDKind:
		return sdktypes.InvalidOrgID, nil
	default:
		return sdktypes.InvalidOrgID, sdkerrors.NewInvalidArgumentError("unhandled id kind %q", id.Kind())
	}
}

func (gdb *gormdb) GetProjectIDOf(ctx context.Context, id sdktypes.ID) (sdktypes.ProjectID, error) {
	var m idFieldNamer

	// Only records that project is mandatory in them.
	switch id.Kind() {
	case sdktypes.ProjectIDKind:
		return sdktypes.FromID[sdktypes.ProjectID](id), nil
	case sdktypes.BuildIDKind:
		m = scheme.Build{}
	case sdktypes.SessionIDKind:
		m = scheme.Session{}
	case sdktypes.ConnectionIDKind:
		m = scheme.Connection{}
	case sdktypes.TriggerIDKind:
		m = scheme.Trigger{}
	case sdktypes.DeploymentIDKind:
		m = scheme.Deployment{}
	case sdktypes.EventIDKind:
		m = scheme.Event{}
	case sdktypes.IntegrationIDKind, sdktypes.OrgIDKind, sdktypes.UserIDKind:
		return sdktypes.InvalidProjectID, nil
	default:
		return sdktypes.InvalidProjectID, sdkerrors.NewInvalidArgumentError("unhandled id kind %q", id.Kind())
	}

	var p struct {
		ProjectID uuid.UUID      `gorm:"column:project_id"`
		Project   scheme.Project `gorm:"foreignKey:project_id"`
	}

	err := gdb.rdb.WithContext(ctx).
		Model(m).
		Where(m.IDFieldName()+" = ?", id.UUIDValue()).
		Select("project_id").
		First(&p).
		Error
	if err != nil {
		return sdktypes.InvalidProjectID, translateError(err)
	}

	return sdktypes.NewIDFromUUID[sdktypes.ProjectID](p.ProjectID), nil
}
