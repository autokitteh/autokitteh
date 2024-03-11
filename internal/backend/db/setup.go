package db

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Setup both sets up the db and also initializes fundemental data.
func Setup(z *zap.Logger, db DB) error {
	ctx := context.Background()

	if err := db.Setup(ctx); err != nil {
		return fmt.Errorf("db.setup: %w", err)
	}

	return nil
}
