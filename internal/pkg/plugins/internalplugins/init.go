package internalplugins

import (
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiplugin"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/pluginimpl"
)

func RegisterAll(f func(apiplugin.PluginName, *pluginimpl.Plugin)) {
	f("http", HTTP)
	f("os", OS)
	f("time", Time)
}
