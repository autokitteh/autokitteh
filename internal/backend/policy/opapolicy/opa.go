package opapolicy

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/contrib/logging/plugins/ozap"
	"github.com/open-policy-agent/opa/v1/sdk"
	sdktest "github.com/open-policy-agent/opa/v1/sdk/test"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/configs/opa_bundles"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/policy"
)

const defaultConfig = "default"

type Config struct {
	// If empty, use bundled config with the name `defaultConfig`.
	// If begings with "!", use bundled config with that name (without the "!" prefix).
	ConfigPath string `koanf:"config_path"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		ConfigPath: "!" + defaultConfig,
	},
}

func startBundleServer(path string) (*sdktest.Server, error) {
	des, err := fs.ReadDir(opa_bundles.FS, path)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	files := make(map[string]string, len(des))

	for _, de := range des {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".rego") || strings.HasSuffix(de.Name(), "_test.rego") {
			continue
		}

		bs, err := fs.ReadFile(opa_bundles.FS, filepath.Join(path, de.Name()))
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}

		files[de.Name()] = string(bs)
	}

	return sdktest.NewServer(sdktest.MockBundle("/bundles/"+path+".tar.gz", files))
}

func startEmbeddedConfig(l *zap.Logger, name string) ([]byte, error) {
	if name == "" {
		name = defaultConfig
	}

	srv, err := startBundleServer(name)
	if err != nil {
		return nil, fmt.Errorf("start bundle server: %w", err)
	}

	l.Warn("using self served embedded opa config", zap.String("name", name), zap.String("url", srv.URL()))

	return []byte(fmt.Sprintf(`
services:
  embedded:
    url: %q
bundles:
  %s:
    resource: /bundles/%s.tar.gz
decision_logs:
  console: true
`, srv.URL(), name, name)), nil
}

func New(cfg *Config, l *zap.Logger) (policy.DecideFunc, error) {
	var (
		cfgf []byte
		err  error
	)

	if strings.HasPrefix(cfg.ConfigPath, "!") {
		if cfgf, err = startEmbeddedConfig(l, cfg.ConfigPath[1:]); err != nil {
			return nil, fmt.Errorf("start embedded config: %w", err)
		}
	} else {
		if cfgf, err = os.ReadFile(cfg.ConfigPath); err != nil {
			return nil, fmt.Errorf("read config file %q: %w", cfg.ConfigPath, err)
		}
	}

	client, err := sdk.New(
		context.Background(),
		sdk.Options{
			Logger:        ozap.Wrap(l, nil),
			ConsoleLogger: ozap.Wrap(l.Named("opaconsole"), nil),
			Config:        bytes.NewReader(cfgf),
			ID:            fixtures.ProcessID(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("new opa client: %w", err)
	}

	return func(ctx context.Context, path string, input any) (any, error) {
		r, err := client.Decision(ctx, sdk.DecisionOptions{
			Path:  path,
			Input: input,
		})
		if err != nil {
			return nil, err
		}

		return r.Result, nil
	}, nil
}
