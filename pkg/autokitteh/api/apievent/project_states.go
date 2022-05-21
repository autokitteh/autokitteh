package apievent

import (
	pbevent "github.com/autokitteh/autokitteh/api/gen/stubs/go/event"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apilang"
)

type projectEventState interface {
	Name() string
	isProjectEventState()
}

//--

type ErrorProjectEventState struct {
	pb *pbevent.ErrorProjectEventState
}

func (*ErrorProjectEventState) isProjectEventState() {}
func (*ErrorProjectEventState) Name() string         { return "ignored" }

func (e *ErrorProjectEventState) RunSummary() *apilang.RunSummary {
	return apilang.MustRunSummaryFromProto(e.pb.RunSummary)
}

func NewErrorProjectEventState(err error, sum *apilang.RunSummary) *ProjectEventState {
	return NewProjectEventState(
		&ErrorProjectEventState{
			pb: &pbevent.ErrorProjectEventState{
				Error:      err.Error(),
				RunSummary: sum.PB(),
			},
		},
	)
}

//--

type IgnoredProjectEventState struct {
	pb *pbevent.IgnoredProjectEventState
}

func (*IgnoredProjectEventState) isProjectEventState() {}
func (*IgnoredProjectEventState) Name() string         { return "ignored" }

func NewIgnoredProjectEventState(reason string) *ProjectEventState {
	return NewProjectEventState(
		&IgnoredProjectEventState{
			pb: &pbevent.IgnoredProjectEventState{
				Reason: reason,
			},
		},
	)
}

//--

type PendingProjectEventState struct {
	pb *pbevent.PendingProjectEventState
}

func (*PendingProjectEventState) isProjectEventState() {}
func (*PendingProjectEventState) Name() string         { return "pending" }

func NewPendingProjectEventState() *ProjectEventState {
	return NewProjectEventState(&PendingProjectEventState{pb: &pbevent.PendingProjectEventState{}})
}

//--

type ProcessingProjectEventState struct {
	pb *pbevent.ProcessingProjectEventState
}

func (*ProcessingProjectEventState) isProjectEventState() {}
func (*ProcessingProjectEventState) Name() string         { return "processing" }

func NewProcessingProjectEventState() *ProjectEventState {
	return NewProjectEventState(
		&ProcessingProjectEventState{
			pb: &pbevent.ProcessingProjectEventState{},
		},
	)
}

//--

type WaitingProjectEventState struct {
	pb *pbevent.WaitingProjectEventState
}

func (*WaitingProjectEventState) isProjectEventState() {}
func (*WaitingProjectEventState) Name() string         { return "waiting" }

func (w *WaitingProjectEventState) RunSummary() *apilang.RunSummary {
	return apilang.MustRunSummaryFromProto(w.pb.RunSummary)
}

func NewWaitingProjectEventState(names []string, sum *apilang.RunSummary) *ProjectEventState {
	return NewProjectEventState(
		&WaitingProjectEventState{
			pb: &pbevent.WaitingProjectEventState{
				Names:      names,
				RunSummary: sum.PB(),
			},
		},
	)
}

//--

type ProcessedProjectEventState struct {
	pb *pbevent.ProcessedProjectEventState
}

func (*ProcessedProjectEventState) isProjectEventState() {}
func (*ProcessedProjectEventState) Name() string         { return "processed" }

func (p *ProcessedProjectEventState) RunSummary() *apilang.RunSummary {
	return apilang.MustRunSummaryFromProto(p.pb.RunSummary)
}

func NewProcessedProjectEventState(sum *apilang.RunSummary) *ProjectEventState {
	return NewProjectEventState(
		&ProcessedProjectEventState{
			pb: &pbevent.ProcessedProjectEventState{RunSummary: sum.PB()},
		},
	)
}
