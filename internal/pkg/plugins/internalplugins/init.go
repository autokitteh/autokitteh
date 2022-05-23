package internalplugins

import (
	"go.autokitteh.dev/sdk/api/apiplugin"
	"go.autokitteh.dev/sdk/pluginimpl"
)

func RegisterAll(f func(apiplugin.PluginName, *pluginimpl.Plugin)) {
	f("http", HTTP)
	f("os", OS)
	f("time", Time)
}
