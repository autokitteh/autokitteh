package experimental

import (
	"errors"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var downCmd = common.StandardCommand(&cobra.Command{
	Use:     "down",
	Short:   "Stop local server",
	Aliases: []string{"d"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("not implemented yet")
	},
})
