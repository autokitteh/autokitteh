/*
This manages autokitteh's configuration and data directories,
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
)

const (
	ConfigEnvVar = "XDG_CONFIG_HOME"
	DataEnvVar   = "XDG_DATA_HOME"

	appName = "autokitteh"
	perm    = 0o700 // drxw------
)

// ConfigHomeDir returns the XDG config-home directory for autokitteh,
// and guarantees that it exists, so callers can use it safely.
func ConfigHomeDir() string { return homeDir(xdg.ConfigHome) }

// DataHomeDir returns the XDG config-home directory for autokitteh,
// and guarantees that it exists, so callers can use it safely.
func DataHomeDir() string { return homeDir(xdg.DataHome) }

func homeDir(baseDir string) string {
	xdg.Reload() // Account for changes in environment variables.

	dir := filepath.Join(baseDir, appName)
	if err := os.MkdirAll(dir, perm); err != nil {
		// An error here is permanent, unfixable automatically, and without a
		// workaround. Examples: no write permission, exists but as a file.
		panic(err)
	}
	return dir
}

func Reload() {
	ConfigHomeDir()
	DataHomeDir()
}
