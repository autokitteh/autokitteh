package types

import (
	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Signal struct {
	ID            uuid.UUID
	WorkflowID    string
	DestinationID sdktypes.EventDestinationID
	Filter        string
}
