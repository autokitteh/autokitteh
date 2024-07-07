package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

var getCmd = common.StandardCommand(&cobra.Command{
	Use:     "get <project name or ID> [--fail]",
	Short:   "Get project details",
	Aliases: []string{"g"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		p, _, err := r.ProjectNameOrID(ctx, args[0])
		if err != nil {
			return err
		}

		if !p.IsValid() {
			return common.FailIfError(cmd, sdkerrors.ErrNotFound, fmt.Sprintf("project <%q>", args[0]))
		}

		common.RenderKVIfV("project", p)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(getCmd)
}
