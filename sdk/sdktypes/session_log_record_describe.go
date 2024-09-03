package sdktypes

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	modulev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/module/v1"
)

type SessionLogRecordDescribeOptions struct {
	IncludeProcessID bool
	IncludeTime      bool
	TimeFormat       string
	Indent           string
	UnwrapValues     bool
}

func (o SessionLogRecordDescribeOptions) indent(w io.Writer) io.Writer {
	indent := o.Indent
	if indent == "" {
		indent = "  "
	}

	return kittehs.NewIndentedStringWriter(w, indent)
}

var DefaultSessionLogRecordDescribeOptions = &SessionLogRecordDescribeOptions{
	IncludeTime:  true,
	TimeFormat:   time.RFC3339,
	Indent:       "  ",
	UnwrapValues: true,
}

func (r SessionLogRecord) ToString() string { return r.Describe(nil) }

func (r SessionLogRecord) Describe(opts *SessionLogRecordDescribeOptions) string {
	if opts == nil {
		opts = DefaultSessionLogRecordDescribeOptions
	}

	m := r.read()

	b := &strings.Builder{}

	if opts.IncludeProcessID && m.ProcessId != "" {
		fmt.Fprintf(b, "%s ", m.ProcessId)
	}

	if opts.IncludeTime && m.T != nil {
		fmt.Fprintf(b, "[%s] ", r.Timestamp().Format(opts.TimeFormat))
	}

	switch {
	case m.Print != nil:
		fmt.Fprintf(b, "Print: %s\n", m.Print.Text)
	case m.StopRequest != nil:
		fmt.Fprintf(b, "Stop Request: %s\n", m.StopRequest.Reason)
	case m.State != nil:
		fmt.Fprint(b, "State: ")
		describeState(opts, b, r.GetState())
	case m.CallSpec != nil:
		fmt.Fprint(b, "Call Spec: ")
		describeCallSpec(opts, b, r.GetCallSpec())
	case m.CallAttemptStart != nil:
		fmt.Fprintf(b, "Call Attempt Start: Attempt #%d\n", m.CallAttemptStart.Num)
	case m.CallAttemptComplete != nil:
		fmt.Fprintln(b, "Call Attempt Complete")
		w := opts.indent(b)
		if m.CallAttemptComplete.IsLast {
			fmt.Fprintln(w, "Final")
		} else {
			fmt.Fprintf(w, "Retry in %s\n", m.CallAttemptComplete.RetryInterval)
		}
		if m.CallAttemptComplete.Result.Error != nil {
			fmt.Fprintln(w, "Error:")
			describeError(opts, opts.indent(w), kittehs.Must1(ProgramErrorFromProto(m.CallAttemptComplete.Result.Error)))
		} else {
			fmt.Fprintln(w, "Return Value: ")
			describeValue(opts, opts.indent(w), kittehs.Must1(ValueFromProto(m.CallAttemptComplete.Result.Value)))
		}
	default:
		fmt.Fprintln(b, "<unknown>")
	}

	return b.String()
}

func describeCallSpec(opts *SessionLogRecordDescribeOptions, w io.Writer, s SessionCallSpec) {
	fmt.Fprintf(w, "seq #%d, %s\n", s.Seq(), s.m.Function.Function.Name)

	w = opts.indent(w)

	fmt.Fprintln(w, "Function:")
	describeFunctionValue(opts, opts.indent(w), s.Function())

	for i, a := range s.m.Args {
		fmt.Fprintf(w, "Arg #%d:\n", i)
		describeValue(opts, opts.indent(w), kittehs.Must1(ValueFromProto(a)))
	}
	for k, v := range s.m.Kwargs {
		fmt.Fprintf(w, "Arg %q: ", k)
		describeValue(opts, opts.indent(w), kittehs.Must1(ValueFromProto(v)))
	}
}

func describeState(opts *SessionLogRecordDescribeOptions, w io.Writer, s SessionState) {
	fmt.Fprintf(w, "%v ", s.Type())

	switch s := s.Concrete().(type) {
	case SessionStateCreated:
		// nop
	case SessionStateStopped:
		fmt.Fprintf(w, "Reason: %s", s.read().Reason)
	case SessionStateError:
		fmt.Fprintln(w, "")
		describeErrorState(opts, opts.indent(w), s)
	case SessionStateRunning:
		if call := s.Call(); call.IsValid() {
			fmt.Fprintln(w, call.GetFunction().Name())
			w := opts.indent(w)
			fmt.Fprintln(w, "Call:")
			describeFunctionValue(opts, opts.indent(w), call)
		} else {
			fmt.Fprintln(w, "")
		}
		fmt.Fprintf(w, "%sRunID: %s\n", opts.Indent, s.read().RunId)
	case SessionStateCompleted:
		fmt.Fprintln(w, "")
		describeCompletedState(opts, opts.indent(w), s)
	default:
		fmt.Fprintln(w, "<unknown>")
	}
}

func describeCompletedState(opts *SessionLogRecordDescribeOptions, w io.Writer, s SessionStateCompleted) {
	m := s.read()

	if len(m.Prints) > 0 {
		fmt.Fprintln(w, "Prints:")
		ww := opts.indent(w)
		for _, p := range m.Prints {
			fmt.Fprintln(ww, p)
		}
	}

	if len(m.Exports) > 0 {
		fmt.Fprintln(w, "Exports:")
		ww := opts.indent(w)
		for k, v := range m.Exports {
			fmt.Fprintf(ww, "%s: ", k)
			describeValue(opts, ww, kittehs.Must1(ValueFromProto(v)))
		}
	}

	if m.ReturnValue != nil {
		fmt.Fprintln(w, "Return Value:")
		describeValue(opts, opts.indent(w), kittehs.Must1(ValueFromProto(m.ReturnValue)))
	}
}

func describeErrorState(opts *SessionLogRecordDescribeOptions, w io.Writer, s SessionStateError) {
	m := s.read()

	if len(m.Prints) > 0 {
		fmt.Fprintln(w, "Prints:")
		ww := opts.indent(w)
		for _, p := range m.Prints {
			fmt.Fprintln(ww, p)
		}
	}

	if m.Error != nil {
		fmt.Fprintln(w, "Error:")
		describeError(opts, opts.indent(w), s.GetProgramError())
	}
}

func describeError(opts *SessionLogRecordDescribeOptions, w io.Writer, e ProgramError) {
	m := e.read()

	if len(m.Extra) != 0 {
		fmt.Fprintln(w, "Extra:")
		ww := opts.indent(w)
		for k, v := range m.Extra {
			fmt.Fprintf(ww, "%s: %s\n", k, v)
		}
	}

	if m.Value != nil {
		fmt.Fprintln(w, "Value:")
		describeValue(opts, opts.indent(w), e.Value())
	}

	if len(m.Callstack) != 0 {
		fmt.Fprintln(w, "Callstack:")
		describeCallstack(opts, opts.indent(w), e.CallStack())
	}
}

func describeCallstack(_ *SessionLogRecordDescribeOptions, w io.Writer, fs []CallFrame) {
	for i, f := range fs {
		fmt.Fprintf(w, "[%d] %s\n", i, f.Location().CanonicalString())
	}
}

func describeFunctionValue(opts *SessionLogRecordDescribeOptions, w io.Writer, v Value) {
	f := v.GetFunction()
	if !f.IsValid() {
		fmt.Fprintln(w, "invalid function")
		return
	}

	fm := f.read()

	var desc string
	if fm.Desc != nil {
		desc = strings.Join(kittehs.Transform(fm.Desc.Input, func(p *modulev1.FunctionField) string {
			s := p.Name
			if p.Type != "" {
				s += ":" + p.Type
			}
			if p.Optional {
				s += "?"
			}
			if p.Kwarg {
				s += "="
			}
			return s
		}), ",")
	}

	fmt.Fprintf(w, "%s(%s)\n", f.Name(), desc)

	if fm.ExecutorId != "" {
		fmt.Fprintf(w, "Executor: %s\n", fm.ExecutorId)
	}

	if len(fm.Flags) != 0 {
		fmt.Fprintf(w, "Flags: %s\n", strings.Join(fm.Flags, ","))
	}

	if len(fm.Data) != 0 {
		fmt.Fprintf(w, "Data: %d bytes\n", len(fm.Data))
	}
}

func describeValue(opts *SessionLogRecordDescribeOptions, w io.Writer, v Value) {
	if opts.UnwrapValues {
		x, err := valueStringUnwrapper.Unwrap(v)
		if err != nil {
			fmt.Fprintf(w, "unwrap: %v\n", err)
			return
		}

		j, err := json.MarshalIndent(x, "", opts.Indent)
		if err != nil {
			fmt.Fprintf(w, "json: %v\n", err)
			return
		}

		fmt.Fprintln(w, string(j))
	} else {
		s, err := v.ToString()
		if err != nil {
			fmt.Println(w, "error: %w", err)
			return
		}

		fmt.Fprintln(w, opts.indent(w), s)
	}
}
