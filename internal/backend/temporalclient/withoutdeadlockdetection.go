package temporalclient

import (
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/workflow"
)

type abuser func()

func (abuser) ToPayload(any) (*commonpb.Payload, error)      { return nil, nil }
func (abuser) FromPayload(*commonpb.Payload, any) error      { return nil }
func (abuser) ToPayloads(...any) (*commonpb.Payloads, error) { return nil, nil }
func (abuser) FromPayloads(*commonpb.Payloads, ...any) error { return nil }
func (abuser) ToString(*commonpb.Payload) string             { return "" }

func (f abuser) ToStrings(*commonpb.Payloads) []string {
	f()
	return nil
}

// WithoutDeadlockDetection allows to execute the `f`, which does not
// perform any temporal operations (ie calling temporal using `wctx`)
// and thus might trigger the temporal deadlock detector.
//
// This function is useful for `f`s that we know are slow and should not
// have their result cached by temporal - they need to run every
// workflow invocation, including during replays. Example usage is resource
// allocation.
//
// See also https://github.com/temporalio/temporal/issues/6546.
// Blame Maxim, not me.
func WithoutDeadlockDetection(wctx workflow.Context, f func()) {
	cvt := workflow.DataConverterWithoutDeadlockDetection(abuser(f))

	// The data converter without deadlock detection is ContextAware,
	// and it must have the workflow context in order to work.
	cvt = cvt.(workflow.ContextAware).WithWorkflowContext(wctx)

	_ = cvt.ToStrings(nil)
}
