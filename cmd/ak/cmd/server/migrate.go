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
		return Setup()
	},
})

func Migrate() error {
	db, err := InitDB(mode)
	if err != nil {
		return err
	}

	if err := db.Migrate(context.Background()); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
