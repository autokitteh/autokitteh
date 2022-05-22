package apievent

import "github.com/autokitteh/idgen"

type EventID string

func (id EventID) String() string { return string(id) }

func NewEventID() EventID { return EventID(idgen.New("E")) }
