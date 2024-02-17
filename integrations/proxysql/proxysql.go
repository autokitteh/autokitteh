package proxysql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/signal18/replication-manager/proxysql"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var hostPortArgs = sdkmodule.WithArgs("host", "port")

var opts = []sdkmodule.Optfn{
	sdkmodule.WithDataFromConfig(func(cfg string) ([]byte, error) {
		if _, err := parseConfig(cfg); err != nil {
			return nil, err
		}

		return []byte(cfg), nil
	}),
	sdkmodule.ExportFunction("add_offline_server", addOfflineServer, hostPortArgs),
	sdkmodule.ExportFunction("add_server", addServer, hostPortArgs),
	sdkmodule.ExportFunction("set_offline", setOffline, hostPortArgs),
	sdkmodule.ExportFunction("set_offline_soft", setOfflineSoft, hostPortArgs),
	sdkmodule.ExportFunction("set_online", setOnline, hostPortArgs),
	sdkmodule.ExportFunction("set_reader", setReader, hostPortArgs),
	sdkmodule.ExportFunction("set_writer", setWriter, hostPortArgs),
	sdkmodule.ExportFunction("add_user", addUser, sdkmodule.WithArgs("user", "password")),
	sdkmodule.ExportFunction("get_hosts_runtime", getHostsRuntime),
	sdkmodule.ExportFunction("get_version", getVersion),
	sdkmodule.ExportFunction("load_servers_to_runtime", loadServersToRuntime),
	sdkmodule.ExportFunction("truncate", truncate),
}

func New() sdkservices.Integration {
	return sdkintegrations.NewIntegration(kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
		IntegrationId: sdktypes.IntegrationIDFromName("proxysql").String(),
		UniqueName:    "proxysql",
		DisplayName:   "ProxySQL",
		Description:   "ProxySQL is an open-source, high performance, high availability, database protocol aware proxy for MySQL.",
		LogoUrl:       "/static/images/proxysql.png",
		UserLinks: map[string]string{
			"1 ProxySQL website": "https://proxysql.com/",
			"2 Rep-Man website":  "https://signal18.io/products/srm",
			"3 Go client API":    "https://pkg.go.dev/github.com/signal18/replication-manager/proxysql",
		},
	})), sdkmodule.New(opts...))
}

func parseConfig(cfg string) (*proxysql.ProxySQL, error) {
	if cfg == "" {
		return nil, errors.New("no config")
	}

	fields := strings.Split(cfg, ";")

	var psql proxysql.ProxySQL

	for _, f := range fields {
		k, v, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("invalid config field: %s", f)
		}

		switch k {
		case "host":
			psql.Host = v
		case "port":
			psql.Port = v
		case "user":
			psql.User = v
		case "password":
			psql.Password = v
		case "writer_hg":
			psql.WriterHG = v
		case "reader_hg":
			psql.ReaderHG = v
		default:
			return nil, fmt.Errorf("unknown config key: %s", k)
		}
	}

	return &psql, nil
}

func connect(ctx context.Context) (*proxysql.ProxySQL, error) {
	data := sdkmodule.FunctionDataFromContext(ctx)
	if data == nil {
		return nil, errors.New("no config data")
	}

	psql, err := parseConfig(string(data))
	if err != nil {
		return nil, err
	}

	if err := psql.Connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return psql, nil
}

func addUser(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var user, password string
	if err := sdkmodule.UnpackArgs(args, kwargs, "user", &user, "password", &password); err != nil {
		return nil, err
	}

	psql, err := connect(ctx)
	if err != nil {
		return nil, err
	}

	if err := psql.AddUser(user, password); err != nil {
		return nil, err
	}

	return sdktypes.NewNothingValue(), nil
}

func getHostsRuntime(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if err := sdkmodule.UnpackArgs(args, kwargs); err != nil {
		return nil, err
	}

	psql, err := connect(ctx)
	if err != nil {
		return nil, err
	}

	rt, err := psql.GetHostsRuntime()
	if err != nil {
		return nil, err
	}

	return sdktypes.NewStringValue(rt), nil
}

func getVersion(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if err := sdkmodule.UnpackArgs(args, kwargs); err != nil {
		return nil, err
	}

	psql, err := connect(ctx)
	if err != nil {
		return nil, err
	}

	return sdktypes.NewStringValue(psql.GetVersion()), nil
}

func loadServersToRuntime(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if err := sdkmodule.UnpackArgs(args, kwargs); err != nil {
		return nil, err
	}

	psql, err := connect(ctx)
	if err != nil {
		return nil, err
	}

	if err := psql.LoadServersToRuntime(); err != nil {
		return nil, err
	}

	return sdktypes.NewNothingValue(), nil
}

func truncate(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if err := sdkmodule.UnpackArgs(args, kwargs); err != nil {
		return nil, err
	}

	psql, err := connect(ctx)
	if err != nil {
		return nil, err
	}

	if err := psql.Truncate(); err != nil {
		return nil, err
	}

	return sdktypes.NewNothingValue(), nil
}

type hostPortFn = func(string, string) error

func hostPortAction(f func(*proxysql.ProxySQL) hostPortFn) sdkexecutor.Function {
	return func(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
		var host, port string
		if err := sdkmodule.UnpackArgs(args, kwargs, "host", &host, "port", &port); err != nil {
			return nil, err
		}

		psql, err := connect(ctx)
		if err != nil {
			return nil, err
		}

		if err := f(psql)(host, port); err != nil {
			return nil, err
		}

		return sdktypes.NewNothingValue(), nil
	}
}

var (
	addOfflineServer = hostPortAction(func(psql *proxysql.ProxySQL) hostPortFn { return psql.AddOfflineServer })
	addServer        = hostPortAction(func(psql *proxysql.ProxySQL) hostPortFn { return psql.AddServer })
	setOffline       = hostPortAction(func(psql *proxysql.ProxySQL) hostPortFn { return psql.SetOffline })
	setOfflineSoft   = hostPortAction(func(psql *proxysql.ProxySQL) hostPortFn { return psql.SetOfflineSoft })
	setOnline        = hostPortAction(func(psql *proxysql.ProxySQL) hostPortFn { return psql.SetOnline })
	setReader        = hostPortAction(func(psql *proxysql.ProxySQL) hostPortFn { return psql.SetReader })
	setWriter        = hostPortAction(func(psql *proxysql.ProxySQL) hostPortFn { return psql.SetWriter })
)
