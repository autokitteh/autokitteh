package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"gopkg.in/yaml.v3"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

const filePermissions = 0o600

var setCmd = common.StandardCommand(&cobra.Command{
	Use:     "set <key> <value>",
	Short:   "Set persistent configuration value",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		// Required to check current configuration is valid
		// and load all possible configuration keys.
		if err := svc.New(common.Config(), svc.RunOptions{Silent: true}).Err(); err != nil {
			filename := common.ConfigYAMLFilePath()
			err = dig.RootCause(err)
			return fmt.Errorf("%q: invalid configuration: %w", filename, err)
		}

		possibleConfigs := common.Config().ListAll()
		if err := validateArgs(args, possibleConfigs); err != nil {
			return err
		}

		cfg, err := currentConfig()
		if err != nil {
			return err
		}

		k, v := args[0], args[1]
		if err := setKeyValue(cfg, k, v); err != nil {
			return err
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}

		if err := validateConfig(data); err != nil {
			return err
		}

		if err := os.WriteFile(common.ConfigYAMLFilePath(), data, filePermissions); err != nil {
			return fmt.Errorf("write file: %w", err)
		}

		return nil
	},
})

func validateArgs(args []string, possibleConfigs []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected exactly 2 arguments, got %d", len(args))
	}

	key := args[0]
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if !slices.Contains(possibleConfigs, key) {
		return fmt.Errorf("key %q is not a valid configuration", key)
	}

	return nil
}

func currentConfig() (map[string]any, error) {
	cfg := make(map[string]any)

	filename := common.ConfigYAMLFilePath()
	data, err := os.ReadFile(filename)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read file: %w", err)
		}
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse file %q: %w", filename, err)
	}

	return cfg, nil
}

func setKeyValue(cfg map[string]any, key, val string) error {
	keyFields := strings.Split(key, svc.ConfigDelim)

	cfgPtr := cfg
	for i, field := range keyFields {
		if i == len(keyFields)-1 {
			cfgPtr[field] = val
			return nil
		}

		if _, ok := cfgPtr[field]; !ok {
			cfgPtr[field] = map[string]any{}
		}
		var ok bool
		cfgPtr, ok = cfgPtr[field].(map[string]any)
		if !ok {
			return fmt.Errorf("key %s is not a valid configuration", key)
		}
	}
	return nil
}

// Validate the new configuration by reloading it from a temporary file.
func validateConfig(data []byte) error {
	temp, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(temp)

	path := filepath.Join(temp, common.ConfigYAMLFileName)
	if err := os.WriteFile(path, data, filePermissions); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	path = os.Getenv(xdg.ConfigEnvVar)
	defer func() {
		if path != "" {
			os.Setenv(xdg.ConfigEnvVar, path)
		} else {
			os.Unsetenv(xdg.ConfigEnvVar)
		}
		xdg.Reload() // Account for change in environment variable.
	}()

	os.Setenv(xdg.ConfigEnvVar, temp)

	if err := common.InitConfig(nil); err != nil {
		return fmt.Errorf("init temp config: %w", err)
	}

	if err := svc.New(common.Config(), svc.RunOptions{Silent: true}).Err(); err != nil {
		err = dig.RootCause(err)
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	return nil
}
