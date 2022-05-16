package eventsrcsstoregorm

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/datatypes"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
)

type eventsrc struct {
	SrcID       string `gorm:"primaryKey"`
	AccountName string
	Enabled     bool
	Types       datatypes.JSON

	CreatedAt time.Time
	UpdatedAt time.Time
}

func decodeEventSource(e *eventsrc) (*apieventsrc.EventSource, error) {
	var updatedAt *time.Time

	if !e.UpdatedAt.IsZero() {
		updatedAt = &e.UpdatedAt
	}

	var types []string

	if e.Types != nil {
		if err := json.Unmarshal(e.Types, &types); err != nil {
			return nil, fmt.Errorf("types: %w", err)
		}
	}

	return apieventsrc.NewEventSource(
		apieventsrc.EventSourceID(e.SrcID),
		(&apieventsrc.EventSourceSettings{}).SetEnabled(e.Enabled).SetTypes(types),
		e.CreatedAt,
		updatedAt,
	)
}

//--

type binding struct {
	SrcID        string `gorm:"primaryKey;index:idx_src_assoc"`
	ProjectID    string `gorm:"primaryKey"`
	Name         string `gorm:"primaryKey"`
	Enabled      bool
	Association  string `gorm:"index:idx_src_assoc"`
	SourceConfig string
	Approved     bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b binding) TableName() string { return "eventsrcs_bindings" }

func decodeBinding(b *binding) (*apieventsrc.EventSourceProjectBinding, error) {
	var updatedAt *time.Time

	if !b.UpdatedAt.IsZero() {
		updatedAt = &b.UpdatedAt
	}

	return apieventsrc.NewEventSourceProjectBinding(
		apieventsrc.EventSourceID(b.SrcID),
		apiproject.ProjectID(b.ProjectID),
		b.Name,
		b.Association,
		b.SourceConfig,
		b.Approved,
		(&apieventsrc.EventSourceProjectBindingSettings{}).
			SetEnabled(b.Enabled),
		b.CreatedAt,
		updatedAt,
	)
}
