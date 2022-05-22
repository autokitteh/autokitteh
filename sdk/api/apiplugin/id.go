package apiplugin

import (
	"fmt"
	"strings"

	"github.com/autokitteh/autokitteh/sdk/api/apiaccount"
)

const sep = "."

type PluginName string

func (n PluginName) String() string { return string(n) }

type PluginID string

func NewPluginID(a apiaccount.AccountName, n PluginName) PluginID {
	return PluginID(fmt.Sprintf("%s%s%s", a.String(), sep, n.String()))
}

func NewInternalPluginID(n PluginName) PluginID {
	return NewPluginID(apiaccount.InternalAccountName, n)
}

func (id PluginID) Empty() bool { return id == "" }

func (id PluginID) String() string { return string(id) }

func (id *PluginID) MaybeString() string {
	if id == nil {
		return ""
	}

	return id.String()
}

func (id PluginID) Split() (apiaccount.AccountName, PluginName) {
	a, b, _ := strings.Cut(id.String(), sep)
	return apiaccount.AccountName(a), PluginName(b)
}

func (id PluginID) AccountName() apiaccount.AccountName {
	n, _ := id.Split()
	return n
}

func (id PluginID) IsInternal() bool { return id.AccountName().IsInternal() }

func (id PluginID) PluginName() PluginName {
	_, n := id.Split()
	return n
}
