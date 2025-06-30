//go:build enterprise
// +build enterprise

package scheme

import (
	"time"

	"github.com/google/uuid"
)

type WorkflowExecutionRequest struct {
	SessionID  uuid.UUID
	WorkflowID string `gorm:"primaryKey"`
	Args       []byte
	Memo       []byte
	AcquiredAt *time.Time
	AcquiredBy *string `gorm:"index:idx_acquired_by_status"` // worker ID that acquired the request

	Status     string    `gorm:"default:'pending';index:idx_acquired_by_status,where:status='in_progress'"`
	CreatedAt  time.Time `gorm:"default:NOW()"`
	RetryCount int       `gorm:"default:0"` // Number of times this request has been retried
}
