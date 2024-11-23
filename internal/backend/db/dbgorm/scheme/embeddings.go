// TODO: Make atlas/goose ignore these structs.
package scheme

import (
	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Embed this to the model to add access control capabilities.
type Owned struct {
	OwnerUserID uuid.UUID `gorm:"index;type:uuid;not null"`

	// TODO: Solve issue with default user.
	// OwnerUser *User `gorm:"foreignKey:OwnerUserID"`
}

func (o Owned) GetOwnerID() sdktypes.OwnerID {
	return sdktypes.NewOwnerID(sdktypes.NewIDFromUUID[sdktypes.UserID](&o.OwnerUserID))
}

// Embed this to the model to denote that it belongs to a project.
type BelongsToProject struct {
	ProjectID uuid.UUID `gorm:"index;type:uuid;not null"`

	Project *Project
}

func (o BelongsToProject) GetProjectID() sdktypes.ProjectID {
	return sdktypes.NewIDFromUUID[sdktypes.ProjectID](&o.ProjectID)
}
