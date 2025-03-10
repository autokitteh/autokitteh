package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListConnectionsFilter struct {
	IntegrationID sdktypes.IntegrationID `json:"integration_id"`
	OrgID         sdktypes.OrgID         `json:"org_id"`
	ProjectID     sdktypes.ProjectID     `json:"project_id"`
	StatusCode    sdktypes.StatusCode    `json:"status_code"`
}

func (f ListConnectionsFilter) AnyIDSpecified() bool {
	return f.IntegrationID.IsValid() || f.OrgID.IsValid() || f.ProjectID.IsValid()
}

type Connections interface {
	Create(ctx context.Context, conn sdktypes.Connection) (sdktypes.ConnectionID, error)
	Delete(ctx context.Context, id sdktypes.ConnectionID) error
	Update(ctx context.Context, conn sdktypes.Connection) error
	List(ctx context.Context, filter ListConnectionsFilter) ([]sdktypes.Connection, error)
	Get(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error)
	Test(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Status, error)
	RefreshStatus(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Status, error)
}
