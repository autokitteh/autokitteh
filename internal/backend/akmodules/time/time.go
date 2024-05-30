// This closely follows https://github.com/google/starlark-go/tree/master/lib/time.
package time

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("time"))

type t0Type time.Time

func New(t0 time.Time) sdkexecutor.Executor {
	return fixtures.NewBuiltinExecutor(
		ExecutorID,
		sdkmodule.ExportValue("nanosecond", sdkmodule.WithValue(sdktypes.NewDurationValue(time.Nanosecond))),
		sdkmodule.ExportValue("microsecond", sdkmodule.WithValue(sdktypes.NewDurationValue(time.Microsecond))),
		sdkmodule.ExportValue("millisecond", sdkmodule.WithValue(sdktypes.NewDurationValue(time.Millisecond))),
		sdkmodule.ExportValue("second", sdkmodule.WithValue(sdktypes.NewDurationValue(time.Second))),
		sdkmodule.ExportValue("minute", sdkmodule.WithValue(sdktypes.NewDurationValue(time.Minute))),
		sdkmodule.ExportValue("hour", sdkmodule.WithValue(sdktypes.NewDurationValue(time.Hour))),

		sdkmodule.ExportFunction("time", newTime, sdkmodule.WithArgs("year?", "month?", "day?", "hour?", "minute?", "second?", "nanosecond?", "location?"), sdkmodule.WithFlag(sdktypes.PureFunctionFlag)),
		sdkmodule.ExportFunction("parse_time", parseTime, sdkmodule.WithArgs("s", "format?", "location?"), sdkmodule.WithFlag(sdktypes.PureFunctionFlag)),
		sdkmodule.ExportFunction("parse_duration", parseDuration, sdkmodule.WithArgs("s"), sdkmodule.WithFlag(sdktypes.PureFunctionFlag)),
		sdkmodule.ExportFunction("from_timestamp", fromTimestamp, sdkmodule.WithArgs("sec", "nsec?"), sdkmodule.WithFlag(sdktypes.PureFunctionFlag)),
		sdkmodule.ExportFunction("is_valid_timezone", isValidTimezone, sdkmodule.WithArgs("tz"), sdkmodule.WithFlag(sdktypes.PureFunctionFlag)),

		// These are non-deterministic.
		// TODO: consider implementing only this as a higher level
		//       module and somehow when `time.now` called in a starlark builtin
		//       library, redirect here.
		sdkmodule.ExportFunction("now", now, sdkmodule.WithFlags(sdktypes.PureFunctionFlag, sdktypes.PrivilidgedFunctionFlag)),
		sdkmodule.ExportFunction("elapsed", t0Type(t0).elapsed, sdkmodule.WithFlags(sdktypes.PureFunctionFlag, sdktypes.PrivilidgedFunctionFlag)),
	)
}

func now(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if err := sdkmodule.UnpackArgs(args, kwargs); err != nil {
		return sdktypes.InvalidValue, err
	}

	wctx := sessioncontext.GetWorkflowContext(ctx)

	var t time.Time

	if err := workflow.SideEffect(wctx, func(workflow.Context) any {
		return time.Now()
	}).Get(&t); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewTimeValue(t), nil
}

func (t0 t0Type) elapsed(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var resolution time.Duration

	if err := sdkmodule.UnpackArgs(args, kwargs, "resolution?", &resolution); err != nil {
		return sdktypes.InvalidValue, err
	}

	wctx := sessioncontext.GetWorkflowContext(ctx)

	var t1 time.Time

	if err := workflow.SideEffect(wctx, func(workflow.Context) any {
		return time.Now()
	}).Get(&t1); err != nil {
		return sdktypes.InvalidValue, err
	}

	delta := t1.Sub(time.Time(t0))
	if resolution != 0 {
		delta = delta.Truncate(resolution)
	}

	return sdktypes.NewDurationValue(delta), nil
}

func newTime(_ context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		year, month, day, hour, min, sec, nsec int
		loc                                    string
	)

	if err := sdkmodule.UnpackArgs(
		args, kwargs,
		"year?", &year,
		"month?", &month,
		"day?", &day,
		"hour?", &hour,
		"minute?", &min,
		"second?", &sec,
		"nanosecond?", &nsec,
		"location?", &loc,
	); err != nil {
		return sdktypes.InvalidValue, err
	}
	if len(args) > 0 {
		return sdktypes.InvalidValue, fmt.Errorf("time: unexpected positional arguments")
	}
	location, err := time.LoadLocation(loc)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	return sdktypes.NewTimeValue(time.Date(year, time.Month(month), day, hour, min, sec, nsec, location)), nil
}

func parseTime(_ context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var s, location, format string
	if err := sdkmodule.UnpackArgs(args, kwargs, "s", &s, "format?", &format, "location?", &location); err != nil {
		return sdktypes.InvalidValue, err
	}

	if location == "" && format == "" {
		var t time.Time

		if err := sdktypes.DefaultValueWrapper.UnwrapInto(&t, args[0]); err != nil {
			return sdktypes.InvalidValue, err
		}

		return sdktypes.NewTimeValue(t), nil
	}

	if location == "" {
		location = "UTC"
	}

	if format == "" {
		format = time.RFC3339
	}

	if location == "UTC" {
		t, err := time.Parse(format, s)
		if err != nil {
			return sdktypes.InvalidValue, err
		}
		return sdktypes.NewTimeValue(t), nil
	}

	loc, err := time.LoadLocation(location)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	t, err := time.ParseInLocation(format, s, loc)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewTimeValue(t), nil
}

func parseDuration(_ context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var d time.Duration
	if err := sdkmodule.UnpackArgs(args, kwargs, "s", &d); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewDurationValue(d), nil
}

func fromTimestamp(_ context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var sec, nsec int64

	if err := sdkmodule.UnpackArgs(args, kwargs, "sec", &sec, "nsec?", &nsec); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewTimeValue(time.Unix(sec, nsec)), nil
}

func isValidTimezone(_ context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var s string
	if err := sdkmodule.UnpackArgs(args, kwargs, "tz", &s); err != nil {
		return sdktypes.InvalidValue, err
	}
	_, err := time.LoadLocation(s)
	return sdktypes.NewBooleanValue(err == nil), nil
}
