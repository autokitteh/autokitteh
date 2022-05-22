package internalplugins

import (
	"github.com/autokitteh/autokitteh/sdk/api/apiplugin"
	"github.com/autokitteh/autokitteh/sdk/pluginimpl"
)

func RegisterAll(f func(apiplugin.PluginName, *pluginimpl.Plugin)) {
	f("http", HTTP)
	f("os", OS)
	f("time", Time)
}
