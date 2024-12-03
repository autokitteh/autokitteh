// TODO: Make atlas/goose ignore these structs.
package scheme

import (
	"time"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Embed this to the model to add access control capabilities.
type Owned struct {
	OwnerID uuid.UUID `gorm:"index;type:uuid;not null"`

	// for foreign key validations.
	OwnerUserID uuid.UUID `gorm:"index;type:uuid;not null"`
	OwnerOrgID  uuid.UUID `gorm:"index;type:uuid;not null"`

	OwnerOrg *Org `gorm:"foreignKey:OwnerOrgID"`

	// TODO: Solve issue with default user.
	// OwnerUser *User `gorm:"foreignKey:OwnerUserID"`
}

func (o Owned) GetOwnerID() sdktypes.OwnerID {
	if o.OwnerUserID != uuid.Nil {
		return sdktypes.NewOwnerID(sdktypes.NewIDFromUUID[sdktypes.UserID](&o.OwnerUserID))
	}

	if o.OwnerOrgID != uuid.Nil {
		return sdktypes.NewOwnerID(sdktypes.NewIDFromUUID[sdktypes.OrgID](&o.OwnerOrgID))
	}

	return sdktypes.InvalidOwnerID
}

// Embed this to the model to denote that it belongs to a project.
type BelongsToProject struct {
	ProjectID uuid.UUID `gorm:"index;type:uuid;not null"`

	Project *Project
}

func (o BelongsToProject) GetProjectID() sdktypes.ProjectID {
	return sdktypes.NewIDFromUUID[sdktypes.ProjectID](&o.ProjectID)
}

// Embed this to the model to add base capabilities.
type Base struct {
	CreatedBy uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time

	// No DeleteAt here as gorm needs it directly in the model
	// in order to recognize it.

	// TODO: Solve issue with default user and org.
	// CreatedByUser *User `gorm:"foreignKey:CreatedBy"`}
	// UpdatedByUser *User `gorm:"foreignKey:UpdatedBy"`}
}
