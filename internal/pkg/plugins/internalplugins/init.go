package internalplugins

import (
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiplugin"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/pluginimpl"
)

func RegisterAll(f func(apiplugin.PluginName, *pluginimpl.Plugin)) {
	f("http", HTTP)
	f("os", OS)
	f("time", Time)
}
