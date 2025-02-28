package common

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/aksvc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

var (
	mode   string
	silent bool
)

// AddModeFlag adds the AutoKitteh service mode flag
// to the given CLI command. See also [ParseModeFlag].
func AddModeFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&mode, "mode", "m", "", "AutoKitteh service mode (default|dev|test)")
}

func AddSilentFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&silent, "silent", "s", false, "Mute all autokitteh logs")
}

// ParseModeFlag returns the parsed value of the AutoKitteh
// service mode CLI flag, initialized with [AddModeFlag].
func ParseModeFlag() (m configset.Mode, err error) {
	if m, err = configset.ParseMode(mode); err != nil {
		err = fmt.Errorf("parse mode: %w", err)
	}
	return
}

// NewDevSvc returns a new AutoKitteh service in dev mode.
func NewDevSvc(silent bool) (aksvc.Service, error) {
	return aksvc.New(Config(), aksvc.RunOptions{Mode: configset.Dev, Silent: silent})
}

// Returns a new AutoKitteh service initialized with the mode
// defined by [AddModeFlag] and parsed by [ParseModeFlag].
func NewSvc() (aksvc.Service, error) {
	m, err := ParseModeFlag()
	if err != nil {
		return nil, err
	}

	return aksvc.New(Config(), aksvc.RunOptions{Mode: m, Silent: silent})
}
