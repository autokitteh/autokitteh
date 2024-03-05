package connections

import (
	"context"
	"errors"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/db"
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
	conn = conn.WithNewID()

	if err := c.DB.CreateConnection(ctx, conn); err != nil {
		return sdktypes.InvalidConnectionID, err
	}

	return conn.ID(), nil
}

func (c *Connections) Update(ctx context.Context, conn sdktypes.Connection) error {
	if err := c.DB.UpdateConnection(ctx, conn); err != nil {
		return err
	}

	return nil
}

func (c *Connections) Delete(ctx context.Context, id sdktypes.ConnectionID) error {
	return c.DB.DeleteConnection(ctx, id)
}

func (c *Connections) Get(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error) {
	desc, err := c.DB.GetConnection(ctx, id)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			return sdktypes.InvalidConnection, nil
		}

		return sdktypes.InvalidConnection, err
	}

	return desc, nil
}

func (c *Connections) List(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
	return c.DB.ListConnections(ctx, filter)
}
