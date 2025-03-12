package sessionworkflows

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// printer is a helper to save session prints to the database.
// It allows for non blocking writes to be performed by calling the
// `Print` method. The saving is done in a separate goroutine, and
// in order to consistently get all written data the `Finalize` method
// must be called.
type printer struct {
	all  []sdkservices.SessionPrint // do not get prints from here, use Finalize() instead.
	ch   chan *print
	done chan struct{}
}

type print struct {
	sdkservices.SessionPrint

	// Correlates a print to an activity call.
	activityCallSeq uint32
}

// Safe to call event if Finalize has been called.
func (p *printer) Close() {
	if p.ch != nil {
		close(p.ch)
	}
}

func (p *printer) Finalize() []sdkservices.SessionPrint {
	p.Close()
	<-p.done
	return p.all
}

func (p *printer) Print(sp sdkservices.SessionPrint, activityCallSeq uint32, isReplay bool) {
	p.all = append(p.all, sp)

	if !isReplay {
		p.ch <- &print{SessionPrint: sp, activityCallSeq: activityCallSeq}
	}
}

func (w *sessionWorkflow) newPrinter() *printer {
	pr := &printer{ch: make(chan *print, 32), done: make(chan struct{})}

	go func() {
		ctx := context.Background()

		for p := range pr.ch {
			if err := w.ws.svcs.DB.AddSessionPrint(ctx, w.data.Session.ID(), p.Value, p.activityCallSeq); err != nil {
				w.l.Error("failed to add session print", zap.Error(err))
			}
		}

		close(pr.done)
		pr.ch = nil
	}()

	return pr
}
