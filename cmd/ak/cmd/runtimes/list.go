package runtimes

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--fail]",
	Short:   "List all registered runtime engines",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		rs, err := common.Client().Runtimes().List(context.Background())
		if err != nil {
			return err
		}

		if kittehs.Must1(cmd.Flags().GetBool("fail")) && len(rs) == 0 {
			return common.NewExitCodeError(common.NotFoundExitCode, errors.New("no runtimes found"))
		}

		common.RenderList(rs)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
