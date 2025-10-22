package opapolicy

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/v1/sdk"
	sdktest "github.com/open-policy-agent/opa/v1/sdk/test"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/configs/opa_bundles"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/policy"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

// Embedded config path in opa_bundles.FS.
const embeddedDefaultConfigPath = "default"

type Config struct {
	ConfigPath         string `koanf:"config_path"`       // if empty, use embedded default config.
	MinLogLevel        string `koanf:"log_level"`         // log level threshold to emit. if empty: "warn".
	MinConsoleLogLevel string `koanf:"console_log_level"` // console log level threshold to emit. if empty: "warn".

	PolicyContentBase64 string `koanf:"policy_content_base64"` // If ConfigPath is empty, and this is set, load policy content from this base64 string.

	// for testing only, to test alternate embedded policies.
	fs fs.FS
}

var Configs = configset.Set[Config]{
	Default: &Config{},
}

// Start an OPA bundle server with bundles configurations stored in bundleFS.
func startBundleServer(bundleFS fs.FS, path string) (*sdktest.Server, error) {
	des, err := fs.ReadDir(bundleFS, path)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	files := make(map[string]string, len(des))

	for _, de := range des {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".rego") || strings.HasSuffix(de.Name(), "_test.rego") {
			continue
		}

		bs, err := fs.ReadFile(bundleFS, filepath.Join(path, de.Name()))
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}

		files[de.Name()] = string(bs)
	}

	return sdktest.NewServer(sdktest.MockBundle("/bundles/"+path+".tar.gz", files))
}

func startEmbeddedConfig(l *zap.Logger, bundleFS fs.FS, name string) ([]byte, error) {
	if name == "" {
		name = embeddedDefaultConfigPath
	}

	srv, err := startBundleServer(bundleFS, name)
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

func parseLogLevel(txt, def string) (zap.AtomicLevel, error) {
	if txt == "" {
		txt = def
	}

	return zap.ParseAtomicLevel(txt)
}

func New(cfg *Config, l *zap.Logger) (policy.DecideFunc, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	var (
		cfgf []byte
		err  error
	)

	if cfg.ConfigPath == "" {
		fs := cfg.fs

		if fs == nil && cfg.PolicyContentBase64 != "" {
			decoded, err := base64.StdEncoding.DecodeString(cfg.PolicyContentBase64)
			if err != nil {
				return nil, fmt.Errorf("decode local policy content: %w", err)
			}

			if fs, err = kittehs.MapToMemFS(map[string][]byte{
				"default/policy.rego": decoded,
			}); err != nil {
				return nil, fmt.Errorf("local policy fs: %w", err)
			}

			l.Warn("using policy content from env var", zap.Int("size_bytes", len(decoded)))
		}

		if fs == nil {
			fs = opa_bundles.FS
		}

		if cfgf, err = startEmbeddedConfig(l, fs, embeddedDefaultConfigPath); err != nil {
			return nil, fmt.Errorf("start embedded config: %w", err)
		}
	} else {
		if cfgf, err = os.ReadFile(cfg.ConfigPath); err != nil {
			return nil, fmt.Errorf("read config file %q: %w", cfg.ConfigPath, err)
		}
	}

	consoleLogLevel, err := parseLogLevel(cfg.MinConsoleLogLevel, "warn")
	if err != nil {
		return nil, fmt.Errorf("parse console log level %q: %w", cfg.MinConsoleLogLevel, err)
	}

	logLevel, err := parseLogLevel(cfg.MinLogLevel, "warn")
	if err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", cfg.MinLogLevel, err)
	}

	client, err := sdk.New(
		context.Background(),
		sdk.Options{
			Logger:        wrapLogger(l.Named("opa"), &logLevel),
			ConsoleLogger: wrapLogger(l.Named("opaconsole"), &consoleLogLevel),
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
