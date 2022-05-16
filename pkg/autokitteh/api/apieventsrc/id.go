package apieventsrc

import (
	"strings"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiaccount"
)

const sep = "."

type EventSourceName string

func (n EventSourceName) String() string { return string(n) }

type EventSourceID string

func (id EventSourceID) Empty() bool { return id == "" }

func (id EventSourceID) String() string { return string(id) }

func (id *EventSourceID) MaybeString() string {
	if id == nil {
		return ""
	}

	return id.String()
}

func (id EventSourceID) Split() (apiaccount.AccountName, EventSourceName) {
	a, b, _ := strings.Cut(id.String(), sep)
	return apiaccount.AccountName(a), EventSourceName(b)
}

func (id EventSourceID) AccountName() apiaccount.AccountName {
	n, _ := id.Split()
	return n
}

func (id EventSourceID) EventSourceName() EventSourceName {
	_, n := id.Split()
	return n
}
