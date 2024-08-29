package connections

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/integrations/webhooks"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type connections struct {
	db           db.DB
	integrations sdkservices.Integrations
	webhooks     *webhooks.Service
}

func New(db db.DB, ints sdkservices.Integrations, wh *webhooks.Service) sdkservices.Connections {
	return &connections{db: db, integrations: ints, webhooks: wh}
}

func (c *connections) Create(ctx context.Context, conn sdktypes.Connection) (sdktypes.ConnectionID, error) {
	intg, err := c.integrations.GetByID(ctx, conn.IntegrationID())
	if err != nil {
		return sdktypes.InvalidConnectionID, err
	}

	if intg == nil {
		return sdktypes.InvalidConnectionID, sdkerrors.ErrNotFound
	}

	status := intg.Get().InitialConnectionStatus()
	if !status.IsValid() {
		// get the connection status of a new connection.
		if status, err = intg.GetConnectionStatus(ctx, sdktypes.InvalidConnectionID); err != nil {
			return sdktypes.InvalidConnectionID, err
		}
	}

	conn = conn.WithStatus(status).WithNewID()

	if err := c.db.CreateConnection(ctx, conn); err != nil {
		return sdktypes.InvalidConnectionID, err
	}

	prev := conn

	if conn, err = c.handleSpecialConnections(ctx, conn); err != nil {
		return sdktypes.InvalidConnectionID, err
	}

	if !prev.Equal(conn) {
		if err := c.db.UpdateConnection(ctx, conn); err != nil {
			return sdktypes.InvalidConnectionID, err
		}
	}

	return conn.ID(), nil
}

func (c *connections) handleSpecialConnections(ctx context.Context, conn sdktypes.Connection) (sdktypes.Connection, error) {
	if conn.IntegrationID() == webhooks.IntegrationID {
		return c.webhooks.ConnectionCreated(ctx, conn)
	}

	return conn, nil
}

func (c *connections) Update(ctx context.Context, conn sdktypes.Connection) error {
	return c.db.UpdateConnection(ctx, conn)
}

func (c *connections) Delete(ctx context.Context, id sdktypes.ConnectionID) error {
	return c.db.DeleteConnection(ctx, id)
}

func (c *connections) List(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
	conns, err := c.db.ListConnections(ctx, filter, false)
	if err != nil {
		return nil, err
	}

	return kittehs.TransformError(conns, func(conn sdktypes.Connection) (sdktypes.Connection, error) {
		return c.enrichConnection(ctx, conn)
	})
}

func (c *connections) getIntegration(ctx context.Context, id sdktypes.ConnectionID) (sdkservices.Integration, error) {
	conn, err := c.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return c.integrations.GetByID(ctx, conn.IntegrationID())
}

func (c *connections) Test(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Status, error) {
	i, err := c.getIntegration(ctx, id)
	if err != nil {
		return sdktypes.InvalidStatus, err
	}

	if i.Get().ConnectionCapabilities().SupportsConnectionTest() {
		return i.TestConnection(ctx, id)
	}

	return sdktypes.InvalidStatus, sdkerrors.ErrNotImplemented
}

func (c *connections) RefreshStatus(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Status, error) {
	i, err := c.getIntegration(ctx, id)
	if err != nil {
		return sdktypes.InvalidStatus, err
	}

	st, err := i.GetConnectionStatus(ctx, id)
	if err != nil {
		return sdktypes.InvalidStatus, err
	}

	if err := c.db.UpdateConnection(ctx, sdktypes.NewConnection(id).WithStatus(st)); err != nil {
		return sdktypes.InvalidStatus, err
	}

	return st, nil
}

func (c *connections) Get(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error) {
	conn, err := c.db.GetConnection(ctx, id)
	if err != nil || !conn.IsValid() {
		return sdktypes.InvalidConnection, err
	}
	return c.enrichConnection(ctx, conn)
}

func (c *connections) enrichConnection(ctx context.Context, conn sdktypes.Connection) (sdktypes.Connection, error) {
	if !conn.IntegrationID().IsValid() && conn.ID() == sdktypes.BuiltinSchedulerConnectionID {
		return conn, nil
	}

	intg, err := c.integrations.GetByID(ctx, conn.IntegrationID())
	if err != nil {
		return sdktypes.InvalidConnection, err
	}

	caps := intg.Get().ConnectionCapabilities()

	// These links are directing to `dashboardsvc`.

	links := map[string]string{
		"self_url": fmt.Sprintf("/connections/%v", conn.ID()),
	}

	if caps.SupportsConnectionInit() {
		links["init_url"] = fmt.Sprintf("/connections/%v/init", conn.ID())
	}

	if caps.SupportsConnectionTest() {
		links["test_url"] = fmt.Sprintf("/connections/%v/test", conn.ID())
	}

	return conn.WithCapabilities(caps).WithLinks(links), nil
}
