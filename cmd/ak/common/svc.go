package common

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

var mode string

func GetMode() (m configset.Mode, err error) {
	if m, err = configset.ParseMode(mode); err != nil {
		err = fmt.Errorf("parse mode: %w", err)
	}
	return
}

func NewDevSvc(silent bool) (svc.Service, error) {
	return svc.New(Config(), svc.RunOptions{Mode: configset.Dev, Silent: silent})
}

func NewSvc(silent bool) (svc.Service, error) {
	m, err := GetMode()
	if err != nil {
		return nil, fmt.Errorf("invalid mode: %w", err)
	}

	return svc.New(Config(), svc.RunOptions{Mode: m, Silent: silent})
}

func AddModeFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&mode, "mode", "m", "", "service mode (default|dev|test)")
}
