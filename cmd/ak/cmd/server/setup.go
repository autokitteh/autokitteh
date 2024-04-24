package server

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var setupCmd = common.StandardCommand(&cobra.Command{
	Use:     "setup",
	Short:   "Set-up local server storage",
	Aliases: []string{"set", "s"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		return Setup()
	},
})

func Setup() error {
	// db, err := InitDB(mode)
	// if err != nil {
	// 	return err
	// }

	// if err := db.Setup(context.Background()); err != nil {
	// 	return fmt.Errorf("setup: %w", err)
	// }

	return nil
}
