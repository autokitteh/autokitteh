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

// type Job struct {
// 	JobID  string `gorm:"primaryKey"`
// 	Data   datatypes.JSON
// 	Status string // Index is created in the migration file manually since gorm doesn't support conditional indexes
// }
