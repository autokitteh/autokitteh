//go:build enterprise
// +build enterprise

package scheme

import (
	"time"

	"github.com/google/uuid"
)

type WorkerInfo struct {
	WorkerID        string `gorm:"primaryKey"`
	ActiveWorkflows int
	UpdatedAt       time.Time
}

type WorkflowExecutionRequest struct {
	SessionID  uuid.UUID `gorm:"primaryKey"`
	Args       []byte
	Memo       []byte
	AcquiredAt *time.Time
	AcquiredBy *string

	CreatedAt time.Time `gorm:"default:NOW()"`
}
