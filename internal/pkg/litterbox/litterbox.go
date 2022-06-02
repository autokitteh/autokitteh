package litterbox

import (
	"context"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apivalues"
)

type LitterBoxID string

type LitterBoxEvent struct {
	SrcBinding string
	Type       string
	Data       map[string]*apivalues.Value
	OriginalID string
}

type LitterBox interface {
	Setup(_ context.Context, id LitterBoxID, sources map[string][]byte, main string) (LitterBoxID, error)
	RunEvent(context.Context, LitterBoxID, *LitterBoxEvent, chan<- *apievent.TrackIngestEventUpdate) error
	Scoop(context.Context, LitterBoxID) error
}
