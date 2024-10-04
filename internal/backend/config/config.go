package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// TODO: need to find a way to list all expected keys.
type Config struct {
	k *koanf.Koanf
}

func (c *Config) Get(path string, dst any) (bool, error) {
	if !c.k.Exists(path) {
		return false, nil
	}

	if err := c.k.Unmarshal(path, dst); err != nil {
		return true, err
	}

	return true, nil
}

// GetConfig returns the unmarshalled configuration in `path`.
// If `path` does not exist, `def` is returned.
// Unmarshalling is done into a prepopulated config with `def`.
func GetConfig[T any](c *Config, path string, def T) (*T, error) {
	dst := def

	if _, err := c.Get(path, &dst); err != nil {
		return nil, err
	}

	return &dst, nil
}

const Delim = "."

// TODO: make sure confmapvs maps to real pre-registered keys.
// If envVarPrefix is empty, do not load from environment.
// If yamlPath is empty, do not load from file.
func LoadConfig(envVarPrefix string, confmapvs map[string]any, yamlPath string) (*Config, error) {
	k := koanf.NewWithConf(koanf.Conf{Delim: Delim, StrictMerge: true})

	// YAML file is loaded first so that env variables can override it.
	if yamlPath != "" {
		err := k.Load(file.Provider(yamlPath), yaml.Parser())
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("load file %q: %w", yamlPath, err)
		}
	}

	if envVarPrefix != "" {
		// Env variables should have the following convention to support hierarchical keys:
		// PREFIX_PARENT1__CHILD1__REQUIRED_KEY_NAME
		// meaning heirarchy is separated by `__` and the name it self use one `_``.
		if err := k.Load(env.Provider(envVarPrefix, "__", func(s string) string {
			return strings.ToLower(strings.TrimPrefix(s, envVarPrefix))
		}), nil); err != nil {
			return nil, fmt.Errorf("load env: %w", err)
		}
	}

	if err := k.Load(confmap.Provider(confmapvs, "."), nil); err != nil {
		return nil, fmt.Errorf("load confmap: %w", err)
	}

	return &Config{k: k}, nil
}

func parseKoanfTags(prefix string, v reflect.Value) []string {
	var result []string

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		tag := fieldType.Tag.Get("koanf")
		if tag == "" {
			continue
		}

		if fieldType.Type.Kind() == reflect.Struct || fieldType.Type.Kind() == reflect.Interface {
			result = append(result, parseKoanfTags(prefix+Delim+tag, field)...)
			continue
		}

		if tag != "" {
			result = append(result, prefix+Delim+tag)
		}
	}

	return result
}
