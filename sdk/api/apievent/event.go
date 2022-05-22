package apievent

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pbevent "github.com/autokitteh/autokitteh/api/gen/stubs/go/event"

	"github.com/autokitteh/autokitteh/sdk/api/apieventsrc"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

type EventPB = pbevent.Event

type Event struct{ pb *pbevent.Event }

func (e *Event) PB() *pbevent.Event {
	if e == nil {
		return nil
	}

	return proto.Clone(e.pb).(*pbevent.Event)
}

func (e *Event) Clone() *Event { return &Event{pb: e.PB()} }

func (e *Event) ID() EventID { return EventID(e.pb.Id) }
func (e *Event) EventSourceID() apieventsrc.EventSourceID {
	return apieventsrc.EventSourceID(e.pb.SrcId)
}
func (e *Event) Memo() map[string]string  { return e.pb.Memo }
func (e *Event) Type() string             { return e.pb.Type }
func (e *Event) OriginalID() string       { return e.pb.OriginalId }
func (e *Event) AssociationToken() string { return e.pb.AssociationToken }
func (e *Event) T() time.Time             { return e.pb.T.AsTime() }

func (e *Event) Data() map[string]*apivalues.Value {
	return apivalues.MustStringValueMapFromProto(e.pb.Data)
}

func (e *Event) AsValue() *apivalues.Value { return apivalues.MustNewValue(e.AsStructValue()) }

func (e *Event) AsStructValue() apivalues.StructValue {
	S := apivalues.String

	data := e.Data()
	var dataDict []*apivalues.DictItem
	for k, v := range data {
		dataDict = append(dataDict, &apivalues.DictItem{K: S(k), V: v})
	}

	return apivalues.StructValue{
		Ctor: apivalues.Symbol("event"),
		Fields: map[string]*apivalues.Value{
			"id":          S(e.ID().String()),
			"type":        S(e.Type()),
			"src_id":      S(e.EventSourceID().String()),
			"original_id": S(e.OriginalID()),
			"t":           apivalues.MustNewValue(apivalues.TimeValue(e.T())),
			"data":        apivalues.MustNewValue(apivalues.DictValue(dataDict)),
		},
	}
}

func EventFromProto(pb *pbevent.Event) (*Event, error) {
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	// TODO: more validation?
	return (&Event{pb: pb}).Clone(), nil
}

func MustNewEvent(
	id EventID,
	srcid apieventsrc.EventSourceID,
	assoc string,
	originalID string,
	typ string,
	data map[string]*apivalues.Value,
	memo map[string]string,
	t time.Time,
) *Event {
	ev, err := NewEvent(id, srcid, assoc, originalID, typ, data, memo, t)
	if err != nil {
		panic(err)
	}
	return ev
}

func NewEvent(
	id EventID,
	srcid apieventsrc.EventSourceID,
	assoc string,
	originalID string,
	typ string,
	data map[string]*apivalues.Value,
	memo map[string]string,
	t time.Time,
) (*Event, error) {
	return EventFromProto(&pbevent.Event{
		Id:               id.String(),
		SrcId:            srcid.String(),
		AssociationToken: assoc,
		Type:             typ,
		Memo:             memo,
		OriginalId:       originalID,
		Data:             apivalues.StringValueMapToProto(data),
		T:                timestamppb.New(t),
	})
}
