package server

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var migrateCmd = common.StandardCommand(&cobra.Command{
	Use:     "migrate",
	Short:   "Update database to latest version",
	Aliases: []string{},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := common.ParseModeFlag()
		if err != nil {
			return err
		}

		db, err := InitDB(common.Config(), mode)
		if err != nil {
			return fmt.Errorf("init DB: %w", err)
		}

		if err := db.Migrate(context.Background()); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}

		return nil
	},
})

func init() {
	common.AddModeFlag(migrateCmd)
}
