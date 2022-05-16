package statestoregorm

import (
	"gorm.io/datatypes"

	"github.com/autokitteh/autokitteh/internal/pkg/statestore"
)

type value struct {
	ProjectID string         `gorm:"primaryKey"`
	Name      string         `gorm:"primaryKey"`
	Value     datatypes.JSON // empty for deleted

	statestore.Metadata
}

func (v value) TableName() string { return "state_values" }
