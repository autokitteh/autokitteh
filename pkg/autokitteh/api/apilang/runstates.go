package apilang

import (
	pblang "github.com/autokitteh/autokitteh/api/gen/stubs/go/lang"
	pbprogram "github.com/autokitteh/autokitteh/api/gen/stubs/go/program"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
)

type runState interface {
	runStateImpl()
	IsRunState() bool // Is an actual state or just an update.
	IsFinal() bool
	IsDiscardable() bool
	Name() string
}

//--

type RunningRunState struct{ pb *pblang.RunningRunState }

func (*RunningRunState) runStateImpl()       {}
func (*RunningRunState) IsRunState() bool    { return true }
func (*RunningRunState) IsFinal() bool       { return false }
func (*RunningRunState) IsDiscardable() bool { return false }
func (*RunningRunState) Name() string        { return "running" }

func NewRunningRunState() *RunState {
	return MustNewRunState(&RunningRunState{pb: &pblang.RunningRunState{}})
}

//--

type CallWaitRunState struct{ pb *pblang.CallWaitRunState }

func (*CallWaitRunState) runStateImpl()       {}
func (*CallWaitRunState) IsRunState() bool    { return true }
func (*CallWaitRunState) IsFinal() bool       { return false }
func (*CallWaitRunState) IsDiscardable() bool { return false }

func (*CallWaitRunState) Name() string { return "call_wait" }

func (c *CallWaitRunState) CallValue() *apivalues.Value {
	return apivalues.MustValueFromProto(c.pb.Call)
}

func (c *CallWaitRunState) Args() []*apivalues.Value {
	return apivalues.MustValuesListFromProto(c.pb.Args)
}

func (c *CallWaitRunState) Kws() map[string]*apivalues.Value {
	return apivalues.MustStringValueMapFromProto(c.pb.Kws)
}

func (c *CallWaitRunState) RunSummary() *RunSummary {
	return MustRunSummaryFromProto(c.pb.RunSummary)
}

func NewCallWaitRunState(call *apivalues.Value, args []*apivalues.Value, kws map[string]*apivalues.Value, sum *RunSummary) *RunState {
	return MustNewRunState(&CallWaitRunState{pb: &pblang.CallWaitRunState{
		Call:       call.PB(),
		Args:       apivalues.ValuesListToProto(args),
		Kws:        apivalues.StringValueMapToProto(kws),
		RunSummary: sum.PB(),
	}})
}

//--

type LoadWaitRunState struct{ pb *pblang.LoadWaitRunState }

func (*LoadWaitRunState) runStateImpl()       {}
func (*LoadWaitRunState) IsRunState() bool    { return true }
func (*LoadWaitRunState) IsFinal() bool       { return false }
func (*LoadWaitRunState) IsDiscardable() bool { return false }
func (*LoadWaitRunState) Name() string        { return "load_wait" }

func (l *LoadWaitRunState) Path() *apiprogram.Path { return apiprogram.MustPathFromProto(l.pb.Path) }

func NewLoadWaitRunState(path *apiprogram.Path) *RunState {
	return MustNewRunState(&LoadWaitRunState{pb: &pblang.LoadWaitRunState{
		Path: path.PB(),
	}})
}

//--

type CompletedRunState struct{ pb *pblang.CompletedRunState }

func (*CompletedRunState) runStateImpl()       {}
func (*CompletedRunState) IsRunState() bool    { return true }
func (*CompletedRunState) IsFinal() bool       { return true }
func (*CompletedRunState) Name() string        { return "completed" }
func (*CompletedRunState) IsDiscardable() bool { return true }

func (c *CompletedRunState) Values() map[string]*apivalues.Value {
	return apivalues.MustStringValueMapFromProto(c.pb.Vals)
}

func NewCompletedRunState(vs map[string]*apivalues.Value) *RunState {
	return MustNewRunState(&CompletedRunState{pb: &pblang.CompletedRunState{
		Vals: apivalues.StringValueMapToProto(vs),
	}})
}

//--

type CanceledRunState struct{ pb *pblang.CanceledRunState }

func (*CanceledRunState) runStateImpl()       {}
func (*CanceledRunState) IsRunState() bool    { return true }
func (*CanceledRunState) IsFinal() bool       { return true }
func (*CanceledRunState) Name() string        { return "canceled" }
func (*CanceledRunState) IsDiscardable() bool { return true }

func (c *CanceledRunState) CallStack() []*apiprogram.CallFrame {
	fs := make([]*apiprogram.CallFrame, len(c.pb.Callstack))
	for i, pbf := range c.pb.Callstack {
		fs[i] = apiprogram.MustCallFrameFromProto(pbf)
	}
	return fs
}

func NewCanceledRunState(reason string, callstack []*apiprogram.CallFrame) *RunState {
	fs := make([]*pbprogram.CallFrame, len(callstack))
	for i, f := range callstack {
		fs[i] = f.PB()
	}

	return MustNewRunState(&CanceledRunState{pb: &pblang.CanceledRunState{
		Reason:    reason,
		Callstack: fs,
	}})
}

//--

type ErrorRunState struct{ pb *pblang.ErrorRunState }

func (*ErrorRunState) runStateImpl()       {}
func (*ErrorRunState) IsRunState() bool    { return true }
func (*ErrorRunState) IsFinal() bool       { return true }
func (*ErrorRunState) Name() string        { return "error" }
func (*ErrorRunState) IsDiscardable() bool { return true }

func (e *ErrorRunState) Error() error { return apiprogram.MustErrorFromProto(e.pb.Error) }

func NewErrorRunState(err *apiprogram.Error) *RunState {
	return MustNewRunState(&ErrorRunState{pb: &pblang.ErrorRunState{
		Error: err.PB(),
	}})
}

//--

type PrintRunUpdate struct{ pb *pblang.PrintRunUpdate }

func (*PrintRunUpdate) runStateImpl()       {}
func (*PrintRunUpdate) IsRunState() bool    { return false }
func (*PrintRunUpdate) IsFinal() bool       { return false }
func (*PrintRunUpdate) Name() string        { return "print" }
func (*PrintRunUpdate) IsDiscardable() bool { return false }

func (r *PrintRunUpdate) Text() string { return r.pb.Text }

func NewPrintRunUpdateState(text string) *RunState {
	return MustNewRunState(&PrintRunUpdate{pb: &pblang.PrintRunUpdate{Text: text}})
}

//--

type CallReturnedRunUpdate struct{ pb *pblang.CallReturnedRunUpdate }

func (*CallReturnedRunUpdate) runStateImpl()       {}
func (*CallReturnedRunUpdate) IsRunState() bool    { return false }
func (*CallReturnedRunUpdate) IsFinal() bool       { return false }
func (*CallReturnedRunUpdate) Name() string        { return "call_ret" }
func (*CallReturnedRunUpdate) IsDiscardable() bool { return false }

func NewCallReturnedRunUpdate(retval *apivalues.Value, err *apiprogram.Error) *RunState {
	return MustNewRunState(&CallReturnedRunUpdate{pb: &pblang.CallReturnedRunUpdate{
		Error:  err.PB(),
		Retval: retval.PB(),
	}})
}

//--

type LoadReturnedRunUpdate struct{ pb *pblang.LoadReturnedRunUpdate }

func (*LoadReturnedRunUpdate) runStateImpl()       {}
func (*LoadReturnedRunUpdate) IsRunState() bool    { return false }
func (*LoadReturnedRunUpdate) IsFinal() bool       { return false }
func (*LoadReturnedRunUpdate) Name() string        { return "load_ret" }
func (*LoadReturnedRunUpdate) IsDiscardable() bool { return false }

func NewLoadReturnedRunUpdate(vs map[string]*apivalues.Value, err *apiprogram.Error, sum *RunSummary) *RunState {
	return MustNewRunState(&LoadReturnedRunUpdate{pb: &pblang.LoadReturnedRunUpdate{
		Error:      err.PB(),
		Vals:       apivalues.StringValueMapToProto(vs),
		RunSummary: sum.PB(),
	}})
}

//--

type ClientErrorRunState struct{ pb *pblang.ClientErrorRunState }

var _ error = &ClientErrorRunState{}

func (*ClientErrorRunState) runStateImpl()       {}
func (*ClientErrorRunState) IsRunState() bool    { return true }
func (*ClientErrorRunState) IsFinal() bool       { return true }
func (*ClientErrorRunState) Name() string        { return "client_error" }
func (*ClientErrorRunState) IsDiscardable() bool { return false }

func (c *ClientErrorRunState) Error() string { return c.pb.Error }

func NewClientErrorRunState(err error) *RunState {
	return MustNewRunState(&ClientErrorRunState{pb: &pblang.ClientErrorRunState{Error: err.Error()}})
}
