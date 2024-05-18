package manifest

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type EffectType string

const (
	NoChange EffectType = "no_change"
	Created  EffectType = "created"
	Updated  EffectType = "updated"
	Deleted  EffectType = "deleted"
)

type Effect struct {
	SubjectID sdktypes.ID
	Type      EffectType
	Text      string
}

type Effects []*Effect

func (e Effects) ProjectIDs() []sdktypes.ProjectID {
	return kittehs.TransformFilter(e, func(effect *Effect) *sdktypes.ProjectID {
		if pid, ok := effect.SubjectID.(sdktypes.ProjectID); ok {
			return &pid
		}
		return nil
	})
}
