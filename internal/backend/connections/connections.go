package connections

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Connections struct {
	fx.In

	Z            *zap.Logger
	DB           db.DB
	Integrations sdkservices.Integrations
}

func New(c Connections) sdkservices.Connections { return &c }

func (c *Connections) Create(ctx context.Context, conn sdktypes.Connection) (sdktypes.ConnectionID, error) {
	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidConnectionID,
		"write:create",
		authz.WithData("connection", conn),
		authz.WithAssociationWithID("integration", conn.IntegrationID()),
		authz.WithAssociationWithID("project", conn.ProjectID()),
	); err != nil {
		return sdktypes.InvalidConnectionID, err
	}

	intg, err := c.Integrations.GetByID(ctx, conn.IntegrationID())
	if err != nil {
		return sdktypes.InvalidConnectionID, err
	}

	if !intg.IsValid() {
		return sdktypes.InvalidConnectionID, sdkerrors.ErrNotFound
	}

	status := intg.InitialConnectionStatus()
	if !status.IsValid() {
		i, err := c.Integrations.Attach(ctx, conn.IntegrationID())
		if err != nil {
			return sdktypes.InvalidConnectionID, err
		}

		// get the connection status of a new connection.
		if status, err = i.GetConnectionStatus(ctx, sdktypes.InvalidConnectionID); err != nil {
			return sdktypes.InvalidConnectionID, err
		}
	}

	conn = conn.WithStatus(status).WithNewID()

	if err := c.DB.CreateConnection(ctx, conn); err != nil {
		return sdktypes.InvalidConnectionID, err
	}

	return conn.ID(), nil
}

func (c *Connections) Update(ctx context.Context, conn sdktypes.Connection) error {
	if err := authz.CheckContext(ctx, conn.ID(), "update:update", authz.WithData("connection", conn)); err != nil {
		return err
	}

	if err := c.DB.UpdateConnection(ctx, conn); err != nil {
		return err
	}

	return nil
}

func (c *Connections) Delete(ctx context.Context, id sdktypes.ConnectionID) error {
	if err := authz.CheckContext(ctx, id, "delete:delete", authz.WithConvertForbiddenToNotFound); err != nil {
		return err
	}

	return c.DB.DeleteConnection(ctx, id)
}

func (c *Connections) List(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
	if !filter.AnyIDSpecified() {
		filter.OrgID = authcontext.GetAuthnInferredOrgID(ctx)
	}

	if err := authz.CheckContext(
		ctx,
		sdktypes.InvalidConnectionID,
		"read:list",
		authz.WithData("filter", filter),
		authz.WithAssociationWithID("org", filter.OrgID),
		authz.WithAssociationWithID("project", filter.ProjectID),
		authz.WithAssociationWithID("integration", filter.IntegrationID),
	); err != nil {
		return nil, err
	}

	conns, err := c.DB.ListConnections(ctx, filter, false)
	if err != nil {
		return nil, err
	}

	return kittehs.TransformError(conns, func(conn sdktypes.Connection) (sdktypes.Connection, error) {
		return c.enrichConnection(ctx, conn)
	})
}

func (c *Connections) attachIntegration(ctx context.Context, id sdktypes.ConnectionID) (sdkservices.Integration, error) {
	conn, err := c.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return c.Integrations.Attach(ctx, conn.IntegrationID())
}

func (c *Connections) Test(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Status, error) {
	if err := authz.CheckContext(ctx, id, "test"); err != nil {
		return sdktypes.InvalidStatus, err
	}

	i, err := c.attachIntegration(ctx, id)
	if err != nil {
		return sdktypes.InvalidStatus, err
	}

	if i.Get().ConnectionCapabilities().SupportsConnectionTest() {
		return i.TestConnection(ctx, id)
	}

	return sdktypes.InvalidStatus, sdkerrors.ErrNotImplemented
}

func (c *Connections) RefreshStatus(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Status, error) {
	if err := authz.CheckContext(ctx, id, "refresh"); err != nil {
		return sdktypes.InvalidStatus, err
	}

	i, err := c.attachIntegration(ctx, id)
	if err != nil {
		return sdktypes.InvalidStatus, err
	}

	st, err := i.GetConnectionStatus(ctx, id)
	if err != nil {
		return sdktypes.InvalidStatus, err
	}

	if err := c.Update(ctx, sdktypes.NewConnection(id).WithStatus(st)); err != nil {
		return sdktypes.InvalidStatus, err
	}

	return st, nil
}

func (c *Connections) Get(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error) {
	if err := authz.CheckContext(ctx, id, "read:get", authz.WithConvertForbiddenToNotFound); err != nil {
		return sdktypes.InvalidConnection, err
	}

	conn, err := c.DB.GetConnection(ctx, id)
	if err != nil || !conn.IsValid() {
		return sdktypes.InvalidConnection, err
	}
	return c.enrichConnection(ctx, conn)
}

func (c *Connections) enrichConnection(ctx context.Context, conn sdktypes.Connection) (sdktypes.Connection, error) {
	intg, err := c.Integrations.GetByID(ctx, conn.IntegrationID())
	if err != nil {
		return sdktypes.InvalidConnection, err
	}

	caps := intg.ConnectionCapabilities()

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
