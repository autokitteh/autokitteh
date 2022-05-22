package pluginsvc

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/exp/maps"

	"google.golang.org/grpc"

	"github.com/autokitteh/autokitteh/internal/app/pluginsgrpcsvc"
	"github.com/autokitteh/autokitteh/internal/pkg/plugin"
	"github.com/autokitteh/autokitteh/internal/pkg/plugin/builtinplugin"

	"github.com/autokitteh/autokitteh/sdk/api/apiplugin"
	"github.com/autokitteh/autokitteh/sdk/pluginimpl"

	"github.com/autokitteh/L"
	"github.com/autokitteh/svc"
)

var (
	pluginID = apiplugin.PluginID(os.Getenv("AK_PLUGIN_ID"))

	ready = struct {
		addr, id string
	}{
		addr: os.Getenv("AK_PROC_READY_ADDRESS"),
		id:   os.Getenv("AK_PROC_READY_ID"),
	}
)

func register(ctx context.Context, l L.L, addr net.Addr) error {
	if ready.id == "" {
		ready.id = fmt.Sprintf("%d", os.Getpid())
	}

	if ready.addr == "" {
		l.Warn("no ready address supplied, not registering")
		return nil
	}

	l.Info("registering", "addr", ready.addr, "code", ready.id)

	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		panic("grpc bind address is not a TCP address")
	}

	resp, err := http.Post(
		fmt.Sprintf("%s?id=%s", ready.addr, url.QueryEscape(ready.id)),
		"application/json",
		bytes.NewBuffer([]byte(tcpAddr.String())),
	)

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response is not OK: %s", resp.Status)
	}

	l.Info("registered")

	return nil
}

func BuildSvcOpts(pls map[apiplugin.PluginID]plugin.Plugin) []svc.OptFunc {
	return []svc.OptFunc{
		svc.WithGRPC(true),
		svc.WithComponent(
			svc.Component{
				Name: "pluginsgrpcsvc",
				Init: func(l L.L) (*pluginsgrpcsvc.Svc, error) {
					l.Info("serving plugins", "plugins", maps.Keys(pls))

					if !pluginID.Empty() && len(pls) != 1 {
						return nil, fmt.Errorf("exactly one plugin must match AK_PLUGIN_ID")
					} else if len(pls) == 0 {
						return nil, fmt.Errorf("no plugins")
					}

					return &pluginsgrpcsvc.Svc{L: L.N(l), Plugins: pls}, nil
				},
				Start: func(srv *grpc.Server, svc *pluginsgrpcsvc.Svc) {
					svc.Register(srv)
				},
			},
		),
		svc.WithReady("ready", register),
	}
}

type Version = svc.Version

func Run(ver *Version, pls ...*pluginimpl.Plugin) {
	svc.SetVersion(ver)

	bipls := make(map[apiplugin.PluginID]plugin.Plugin, len(pls))

	for _, pl := range pls {
		var plid apiplugin.PluginID

		if pluginID.Empty() {
			plid = apiplugin.PluginID(pl.ID)
		} else if pl.ID == "" || pl.ID == pluginID.PluginName().String() {
			plid = pluginID
		} else {
			continue
		}

		bipls[plid] = &builtinplugin.BuiltinPlugin{Plugin: pl, ID: plid}
	}

	svc.RunCLI("", BuildSvcOpts(bipls)...)
}
