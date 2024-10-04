package common

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/backend/config"
)

var mode string

// AddModeFlag adds the AutoKitteh service mode flag
// to the given CLI command. See also [ParseModeFlag].
func AddModeFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&mode, "mode", "m", "", "AutoKitteh service mode (default|dev|test)")
}

// ParseModeFlag returns the parsed value of the AutoKitteh
// service mode CLI flag, initialized with [AddModeFlag].
func ParseModeFlag() (m config.Mode, err error) {
	if m, err = config.ParseMode(mode); err != nil {
		err = fmt.Errorf("parse mode: %w", err)
	}
	return
}

func NewDevSvcOpts(silent bool) ([]fx.Option, error) {
	opts := svc.NewFXOpts(Config(), svc.RunOptions{Mode: config.Dev, Silent: silent})
	if err := fx.ValidateApp(opts...); err != nil {
		return nil, err
	}

	return opts, nil
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
