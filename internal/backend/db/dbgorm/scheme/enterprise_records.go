//go:build enterprise
// +build enterprise

package scheme

import (
	"time"
)

type WorkerInfo struct {
	WorkerID        string `gorm:"primaryKey"`
	ActiveWorkflows int
	UpdatedAt       time.Time
}

// type WorkflowExecutionRequest struct {
// 	RequestID  string `gorm:"primaryKey"`
// 	SessionID  string
// 	AqcuiredAt *time.Time
// 	AqcuiredBy string
// }
