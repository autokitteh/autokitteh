package apievent

import (
	pbevent "github.com/autokitteh/autokitteh/api/gen/stubs/go/event"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
)

type eventState interface {
	Name() string
	isEventState()
}

//--

type ErrorEventState struct{ pb *pbevent.ErrorEventState }

func (*ErrorEventState) isEventState() {}
func (*ErrorEventState) Name() string  { return "error" }

func NewErrorEventState(err error) *EventState {
	return NewEventState(&ErrorEventState{pb: &pbevent.ErrorEventState{Error: err.Error()}})
}

//--

type IgnoredEventState struct{ pb *pbevent.IgnoredEventState }

func (*IgnoredEventState) isEventState() {}
func (*IgnoredEventState) Name() string  { return "ignored" }

func NewIgnoredEventState(reason string) *EventState {
	return NewEventState(&IgnoredEventState{pb: &pbevent.IgnoredEventState{Reason: reason}})
}

//--

type PendingEventState struct{ pb *pbevent.PendingEventState }

func (*PendingEventState) isEventState() {}
func (*PendingEventState) Name() string  { return "pending" }

func NewPendingEventState() *EventState {
	return NewEventState(&PendingEventState{pb: &pbevent.PendingEventState{}})
}

//--

type ProcessingEventState struct{ pb *pbevent.ProcessingEventState }

func (*ProcessingEventState) isEventState() {}
func (*ProcessingEventState) Name() string  { return "processing" }

func (p *ProcessingEventState) ProjectIDs() (ret []apiproject.ProjectID) {
	for _, pid := range p.pb.ProjectIds {
		ret = append(ret, apiproject.ProjectID(pid))
	}
	return
}

func (p *ProcessingEventState) IgnoredProjectIDs() (ret []apiproject.ProjectID) {
	for _, pid := range p.pb.IgnoredProjectIds {
		ret = append(ret, apiproject.ProjectID(pid))
	}
	return
}

func NewProcessingEventState(pids, ignoredpids []apiproject.ProjectID) *EventState {
	pbpids := make([]string, len(pids))
	for i, pid := range pids {
		pbpids[i] = pid.String()
	}

	pbignoredpids := make([]string, len(ignoredpids))
	for i, pid := range ignoredpids {
		pbignoredpids[i] = pid.String()
	}

	return NewEventState(&ProcessingEventState{pb: &pbevent.ProcessingEventState{
		ProjectIds:        pbpids,
		IgnoredProjectIds: pbignoredpids,
	}})
}

//--

type ProcessedEventState struct{ pb *pbevent.ProcessedEventState }

func (*ProcessedEventState) isEventState() {}
func (*ProcessedEventState) Name() string  { return "processed" }

func (p *ProcessedEventState) ProjectIDs() (ret []apiproject.ProjectID) {
	for _, pid := range p.pb.ProjectIds {
		ret = append(ret, apiproject.ProjectID(pid))
	}
	return
}

func (p *ProcessedEventState) AttnProjectIDs() (ret []apiproject.ProjectID) {
	for _, pid := range p.pb.AttnProjectIds {
		ret = append(ret, apiproject.ProjectID(pid))
	}
	return
}

func NewProcessedEventState(pids, attn []apiproject.ProjectID) *EventState {
	pbpids := make([]string, len(pids))
	for i, pid := range pids {
		pbpids[i] = pid.String()
	}

	strs := make([]string, len(attn))
	for i, pid := range attn {
		strs[i] = pid.String()
	}

	return NewEventState(&ProcessedEventState{pb: &pbevent.ProcessedEventState{ProjectIds: pbpids, AttnProjectIds: strs}})
}
