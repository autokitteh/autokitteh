package sessionworkflows

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const maxPrints = 1024

// printer is a helper to save session prints to the database.
// It allows for non blocking writes to be performed by calling the
// `Print` method. The saving is done in a separate goroutine, and
// in order to consistently get all written data the `Finalize` method
// must be called.
type printer struct {
	l   *zap.Logger
	db  db.DB
	sid sdktypes.SessionID

	all      []sdkservices.SessionPrint // do not get prints from here, use Finalize() instead.
	overflow bool                       // true if exceeded max prints.

	ch   chan *print   // prints queue to be saved.
	done chan struct{} // closed when printer needs to shut down.
}

type print struct {
	sdkservices.SessionPrint

	// Correlates a print to an activity call.
	activityCallSeq uint32
}

// Must be called in order to stop the printer goroutine, otherwise
// will leak them.
func (p *printer) Finalize() []sdkservices.SessionPrint {
	close(p.ch)
	<-p.done
	return p.all
}

func (p *printer) Print(sp sdkservices.SessionPrint, activityCallSeq uint32, isReplay bool) {
	p.all = append(p.all, sp)

	if len(p.all) > maxPrints {
		if !p.overflow {
			p.l.Warn("too many prints", zap.Int("max", maxPrints))
		}

		p.overflow = true
		p.all = p.all[1:]
	}

	if !isReplay {
		p.ch <- &print{SessionPrint: sp, activityCallSeq: activityCallSeq}
	}
}

func (w *sessionWorkflow) newPrinter() *printer {
	pr := &printer{ch: make(chan *print, 32), done: make(chan struct{}), l: w.l, db: w.ws.svcs.DB, sid: w.data.Session.ID()}

	return pr
}

func (pr *printer) Start() {
	go func() {
		ctx := context.Background()

		for p := range pr.ch {
			if err := pr.db.AddSessionPrint(ctx, pr.sid, p.Value, p.activityCallSeq); err != nil {
				pr.l.Error("failed to add session print", zap.Error(err))
			}
		}

		close(pr.done)
		pr.ch = nil
	}()
}
