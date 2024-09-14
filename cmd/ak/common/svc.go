package common

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var mode string

// AddModeFlag adds the AutoKitteh service mode flag
// to the given CLI command. See also [ParseModeFlag].
func AddModeFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&mode, "mode", "m", "", "AutoKitteh service mode (default|dev|test)")
}

// ParseModeFlag returns the parsed value of the AutoKitteh
// service mode CLI flag, initialized with [AddModeFlag].
func ParseModeFlag() (m configset.Mode, err error) {
	if m, err = configset.ParseMode(mode); err != nil {
		err = kittehs.ErrorWithPrefix("parse mode", err)
	}
	return
}

// NewDevSvc returns a new AutoKitteh service in dev mode.
func NewDevSvc(silent bool) (svc.Service, error) {
	return svc.New(Config(), svc.RunOptions{Mode: configset.Dev, Silent: silent})
}

// Returns a new AutoKitteh service initialized with the mode
// defined by [AddModeFlag] and parsed by [ParseModeFlag].
func NewSvc(silent bool) (svc.Service, error) {
	m, err := ParseModeFlag()
	if err != nil {
		return nil, err
	}

	return svc.New(Config(), svc.RunOptions{Mode: m, Silent: silent})
}
