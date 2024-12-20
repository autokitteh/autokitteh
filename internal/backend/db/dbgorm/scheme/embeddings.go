// TODO: Make atlas/goose ignore these structs.
package scheme

import (
	"time"

	"github.com/google/uuid"
)

// Embed this to the model to add base capabilities.
type Base struct {
	CreatedBy uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time

	// No DeleteAt here as gorm needs it directly in the model
	// in order to recognize it.
}
