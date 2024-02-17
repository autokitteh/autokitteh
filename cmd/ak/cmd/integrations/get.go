package integrations

import (
	"errors"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <integration name or ID> [--fail]",
	Short:   "Get integration details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		i, _, err := r.IntegrationNameOrID(args[0])
		if err != nil {
			if errors.As(err, resolver.NotFoundErrorType) {
				err = common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			return err
		}

		if err := common.FailIfNotFound(cmd, "integration", i); err != nil {
			return err
		}

		common.RenderKVIfV("integration", i)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
