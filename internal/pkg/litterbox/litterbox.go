package litterbox

import (
	"context"
	"errors"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apivalues"
)

var (
	ErrNoSources        = errors.New("no sources supplied")
	ErrMainNotSpecified = errors.New("main not specified")
	ErrNotFound         = errors.New("not found")
)

type LitterBoxID string

type LitterBoxEvent struct {
	Src        string
	Type       string
	Data       map[string]*apivalues.Value
	OriginalID string
}

type LitterBox interface {
	Setup(_ context.Context, id LitterBoxID, files []byte) (LitterBoxID, error)
	RunEvent(context.Context, LitterBoxID, *LitterBoxEvent, chan<- *apievent.TrackIngestEventUpdate) error
	Run(context.Context, LitterBoxID, chan<- *apievent.TrackIngestEventUpdate) error
	Get(context.Context, LitterBoxID) ([]byte, error)
	Scoop(context.Context, LitterBoxID) error
}
