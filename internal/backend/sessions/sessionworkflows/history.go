package sessionworkflows

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type reverseIter struct {
	rs  []*historypb.HistoryEvent
	err error
}

func (i *reverseIter) HasNext() bool { return i.err == nil && len(i.rs) > 0 }

func (i *reverseIter) Next() (*historypb.HistoryEvent, error) {
	if i.err != nil {
		return nil, i.err
	}

	if len(i.rs) == 0 {
		return nil, nil
	}

	e := i.rs[0]
	i.rs = i.rs[1:]

	return e, nil
}

func (ws *workflows) GetWorkflowLog(ctx context.Context, filter sdkservices.SessionLogRecordsFilter) (*sdkservices.GetLogResults, error) {
	l := ws.l.With(zap.String("session_id", filter.SessionID.String()))

	var rs []sdktypes.SessionLogRecord

	iter := ws.svcs.Temporal().GetWorkflowHistory(
		ctx,
		workflowID(filter.SessionID),
		"",
		false,
		enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT,
	)

	if !filter.Ascending {
		// This is an ugly hack as it consumes all elements, but that's what we can do right now.

		riter := &reverseIter{}

		for iter.HasNext() {
			event, err := iter.Next()
			if err != nil {
				riter = &reverseIter{err: err}
				break
			}

			riter.rs = append(riter.rs, event)
		}

		slices.Reverse(riter.rs)

		iter = riter
	}

	var (
		count         int64
		nextPageToken string
		reached       = filter.PageToken == ""
	)

	events := make(map[int64]*historypb.HistoryEvent, min(16, filter.PageSize))

	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			var notFound *serviceerror.NotFound
			if errors.As(err, &notFound) {
				// lost.
				return nil, sdkerrors.ErrNotFound
			}

			return nil, temporalclient.TranslateError(err, "get workflow history")
		}

		if event == nil {
			l.Error("nil event from temporal")
			continue
		}

		var t time.Time

		if pbt := event.GetEventTime(); pbt != nil {
			t = pbt.AsTime()
		} else {
			l.Warn("nil event time from temporal")
		}

		eid := event.EventId

		l := l.With(zap.Int64("event_id", eid))

		events[eid] = event

		r, err := parseTemporalHistoryEvent(l, t, event, events, filter.Types)
		if err != nil {
			return nil, fmt.Errorf("event %d: %w", event.GetEventId(), err)
		}

		if !r.IsValid() {
			continue
		}

		count++

		if nextPageToken != "" || count <= int64(filter.Skip) {
			continue
		}

		if !reached {
			if strconv.FormatInt(eid, 16) == filter.PageToken {
				reached = true
			}

			continue
		}

		if filter.PageSize > 0 && len(rs) >= int(filter.PageSize) {
			nextPageToken = strconv.FormatInt(eid, 16)
			break
		}

		rs = append(rs, r)
	}

	return &sdkservices.GetLogResults{
		Records: rs,
		PaginationResult: sdktypes.PaginationResult{
			TotalCount:    count,
			NextPageToken: nextPageToken,
		},
	}, nil
}

func parseTemporalHistoryEvent(l *zap.Logger, t time.Time, event *historypb.HistoryEvent, events map[int64]*historypb.HistoryEvent, types sdktypes.SessionLogRecordType) (sdktypes.SessionLogRecord, error) {
	switch a := event.Attributes.(type) {
	case *historypb.HistoryEvent_WorkflowExecutionStartedEventAttributes:
		if types != 0 && types&sdktypes.StateSessionLogRecordType == 0 {
			return sdktypes.InvalidSessionLogRecord, nil
		}

		return sdktypes.NewStateSessionLogRecord(t, sdktypes.NewSessionStateCreated()), nil

	case *historypb.HistoryEvent_WorkflowExecutionCancelRequestedEventAttributes:
		if types != 0 && types&sdktypes.StopRequestSessionLogRecordType == 0 {
			return sdktypes.InvalidSessionLogRecord, nil
		}

		a, ok := event.Attributes.(*historypb.HistoryEvent_WorkflowExecutionCancelRequestedEventAttributes)
		if !ok {
			return sdktypes.InvalidSessionLogRecord, errors.New("not a cancel event")
		}

		if a == nil {
			l.Error("nil activity task cancel event attributes from temporal")
			return sdktypes.InvalidSessionLogRecord, nil
		}

		return sdktypes.NewStopRequestSessionLogRecord(t, a.WorkflowExecutionCancelRequestedEventAttributes.GetCause()), nil

	case *historypb.HistoryEvent_WorkflowExecutionTerminatedEventAttributes:
		if types != 0 && types&sdktypes.StateSessionLogRecordType == 0 {
			return sdktypes.InvalidSessionLogRecord, nil
		}

		a, ok := event.Attributes.(*historypb.HistoryEvent_WorkflowExecutionTerminatedEventAttributes)
		if !ok {
			return sdktypes.InvalidSessionLogRecord, errors.New("not a terminate event")
		}

		if a == nil {
			l.Error("nil activity task terminate event attributes from temporal")
			return sdktypes.InvalidSessionLogRecord, nil
		}

		return sdktypes.NewStateSessionLogRecord(t, sdktypes.NewSessionStateStopped("[forced] "+a.WorkflowExecutionTerminatedEventAttributes.Reason)), nil

	case *historypb.HistoryEvent_ActivityTaskScheduledEventAttributes:
		a, ok := event.Attributes.(*historypb.HistoryEvent_ActivityTaskScheduledEventAttributes)
		if !ok {
			return sdktypes.InvalidSessionLogRecord, errors.New("not a scheduled activity event")
		}

		if a == nil {
			l.Error("nil activity task scheduled event attributes from temporal")
			return sdktypes.InvalidSessionLogRecord, nil
		}

		attrs := a.ActivityTaskScheduledEventAttributes
		if attrs == nil {
			l.Error("nil activity task scheduled event attributes from temporal")
			return sdktypes.InvalidSessionLogRecord, nil
		}

		switch attrs.ActivityType.GetName() {
		case sessioncalls.CallActivityName:
			if types != 0 && types&sdktypes.CallSpecSessionLogRecordType == 0 {
				return sdktypes.InvalidSessionLogRecord, nil
			}

			payloads := attrs.Input.GetPayloads()
			if len(payloads) != 1 {
				l.Error("unexpected number of payloads for activity task scheduled event", zap.Int("payloads", len(payloads)))
				return sdktypes.InvalidSessionLogRecord, nil
			}

			var callInputs sessioncalls.CallActivityInputs
			if err := json.Unmarshal(payloads[0].GetData(), &callInputs); err != nil {
				return sdktypes.InvalidSessionLogRecord, temporalclient.TranslateError(err, "unmarshal activity input")
			}

			return sdktypes.NewCallSpecSessionLogRecord(t, callInputs.CallSpec), nil

		case updateSessionStateActivityName:
			if types != 0 && types&sdktypes.StateSessionLogRecordType == 0 {
				return sdktypes.InvalidSessionLogRecord, nil
			}

			payloads := attrs.Input.GetPayloads()
			if len(payloads) != 2 {
				l.Error("unexpected number of payloads for activity task scheduled event", zap.Int("payloads", len(payloads)))
				return sdktypes.InvalidSessionLogRecord, nil
			}

			var state sdktypes.SessionState
			if err := json.Unmarshal(payloads[1].GetData(), &state); err != nil {
				return sdktypes.InvalidSessionLogRecord, temporalclient.TranslateError(err, "unmarshal session state")
			}

			return sdktypes.NewStateSessionLogRecord(t, state), nil

		case addSessionPrintActivityName:
			if types != 0 && types&sdktypes.PrintSessionLogRecordType == 0 {
				return sdktypes.InvalidSessionLogRecord, nil
			}

			payloads := attrs.Input.GetPayloads()
			if len(payloads) != 3 {
				l.Error("unexpected number of payloads for activity task print event", zap.Int("payloads", len(payloads)))
				return sdktypes.InvalidSessionLogRecord, nil
			}

			var v sdktypes.Value
			if err := v.UnmarshalJSON(payloads[1].GetData()); err != nil {
				return sdktypes.InvalidSessionLogRecord, temporalclient.TranslateError(err, "unmarshal print text")
			}

			return sdktypes.NewPrintSessionLogRecord(t, v, 0), nil

		default:
			return sdktypes.InvalidSessionLogRecord, nil
		}

	case *historypb.HistoryEvent_ActivityTaskStartedEventAttributes:
		if types != 0 && types&sdktypes.CallAttemptStartSessionLogRecordType == 0 {
			return sdktypes.InvalidSessionLogRecord, nil
		}

		attrs := a.ActivityTaskStartedEventAttributes
		if attrs == nil {
			l.Error("nil activity task started event attributes from temporal")
			return sdktypes.InvalidSessionLogRecord, nil
		}

		if !wasRelevantEventCall(events, attrs.GetScheduledEventId()) {
			return sdktypes.InvalidSessionLogRecord, nil
		}

		return sdktypes.NewCallAttemptStartSessionLogRecord(
			t,
			sdktypes.NewSessionCallAttemptStart(t, uint32(attrs.Attempt)),
		), nil

	case *historypb.HistoryEvent_ActivityTaskCompletedEventAttributes:
		if types != 0 && types&sdktypes.CallAttemptCompleteSessionLogRecordType == 0 {
			return sdktypes.InvalidSessionLogRecord, nil
		}

		attrs := a.ActivityTaskCompletedEventAttributes
		if attrs == nil {
			l.Error("nil activity task completed event attributes from temporal")
			return sdktypes.InvalidSessionLogRecord, nil
		}

		if !wasRelevantEventCall(events, attrs.GetScheduledEventId()) {
			return sdktypes.InvalidSessionLogRecord, nil
		}

		payloads := attrs.Result.GetPayloads()
		if len(payloads) != 1 {
			l.Error("unexpected number of payloads for activity task completed event", zap.Int("payloads", len(payloads)))
			return sdktypes.InvalidSessionLogRecord, nil
		}

		var callOutputs sessioncalls.CallActivityOutputs

		if err := json.Unmarshal(payloads[0].GetData(), &callOutputs); err != nil {
			return sdktypes.InvalidSessionLogRecord, temporalclient.TranslateError(err, "unmarshal activity input")
		}

		return sdktypes.NewCallAttemptCompleteSessionLogRecord(
			t,
			sdktypes.NewSessionCallAttemptComplete(
				t,
				true,
				callOutputs.Result,
			),
		), nil

	case *historypb.HistoryEvent_ActivityTaskCanceledEventAttributes:
		if types != 0 && types&sdktypes.CallAttemptCompleteSessionLogRecordType == 0 {
			return sdktypes.InvalidSessionLogRecord, nil
		}

		attrs := a.ActivityTaskCanceledEventAttributes
		if attrs == nil {
			l.Error("nil activity task cancelled event attributes from temporal")
			return sdktypes.InvalidSessionLogRecord, nil
		}

		if !wasRelevantEventCall(events, attrs.GetScheduledEventId()) {
			return sdktypes.InvalidSessionLogRecord, nil
		}

		return sdktypes.NewCallAttemptCompleteSessionLogRecord(
			t,
			sdktypes.NewSessionCallAttemptComplete(
				t,
				true,
				sdktypes.NewSessionCallAttemptResult(
					sdktypes.InvalidValue,
					workflow.ErrCanceled,
				),
			),
		), nil

	case *historypb.HistoryEvent_ActivityTaskFailedEventAttributes, *historypb.HistoryEvent_ActivityTaskTimedOutEventAttributes:
		// do not report this, as it is always for infrastructure failures.
		// we retry these indefinitely, so they are not interesting.
		return sdktypes.InvalidSessionLogRecord, nil

	default:
		return sdktypes.InvalidSessionLogRecord, nil
	}
}

func wasRelevantEventCall(events map[int64]*historypb.HistoryEvent, eid int64) bool {
	event, ok := events[eid]
	if !ok {
		return false
	}

	a, ok := event.Attributes.(*historypb.HistoryEvent_ActivityTaskScheduledEventAttributes)
	if !ok || a == nil {
		return false
	}

	attrs := a.ActivityTaskScheduledEventAttributes

	return attrs != nil && attrs.ActivityType.Name == sessioncalls.CallActivityName
}
