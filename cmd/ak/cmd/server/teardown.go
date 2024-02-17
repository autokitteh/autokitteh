package server

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var teardownCmd = common.StandardCommand(&cobra.Command{
	Use:     "teardown",
	Short:   "Tear-down local server storage",
	Aliases: []string{"tear", "td", "t"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		return Teardown()
	},
})

func Teardown() error {
	db, err := InitDB(mode)
	if err != nil {
		return err
	}

	if err := db.Teardown(context.Background()); err != nil {
		return fmt.Errorf("teardown: %w", err)
	}

	return nil
}
