/*
Package XDG manages autokitteh's configuration and data directories,
which are used to store optional files such as ".env", "config.yaml"
(see the CLI command "ak config"), "fake_secrets_manager.json" (if you
opt-out of using a real secrets manager), SaaS client credentials, etc.

This implementation obeys the XDG Base Directory Specification:
https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html

The common exceptions to the default base directories are:

  - macOS (a.k.a. Darwin)
  - Plan 9
  - Windows
*/
package xdg

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

const (
	ConfigEnvVar = "XDG_CONFIG_HOME"
	DataEnvVar   = "XDG_DATA_HOME"

	appName = "autokitteh"
	perm    = 0o700 // drxw------
)

// ConfigHomeDir returns the XDG config-home directory for autokitteh,
// and guarantees that it exists, so callers can use it safely.
func ConfigHomeDir() string { config, _ := dirs(); return config }

// DataHomeDir returns the XDG config-home directory for autokitteh,
// and guarantees that it exists, so callers can use it safely.
func DataHomeDir() string { _, data := dirs(); return data }

func homeDir(baseDir string) string { return filepath.Join(baseDir, appName) }

func ensure(path string) {
	// An error here is permanent, unfixable automatically, and without a
	// workaround. Examples: no write permission, exists but as a file.
	kittehs.Must0(os.MkdirAll(path, perm))
}

func dirs() (config, data string) {
	xdg.Reload() // Account for changes in environment variables.

	// These are used in github.com/adrg/xdg to explicitly set the paths.
	_, explicit := os.LookupEnv(ConfigEnvVar)
	if !explicit {
		_, explicit = os.LookupEnv(DataEnvVar)
	}

	config, data = homeDir(xdg.ConfigHome), homeDir(xdg.DataHome)
	if config == data && !explicit {
		// Both paths are the same, we better separate them.
		config = filepath.Join(config, "config")
		data = filepath.Join(data, "data")
	}

	ensure(config)
	ensure(data)

	return
}

func Reload() { _, _ = dirs() }
