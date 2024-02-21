package builds

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [project name or ID] [--fail]",
	Short:   "List all uploaded project builds",
	Aliases: []string{"ls", "l"},
	Args:    cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			args = append(args, "")
		}

		if args[0] != "" {
			r := resolver.Resolver{Client: common.Client()}
			p, _, err := r.ProjectNameOrID(args[0])
			if err != nil {
				return err
			}
			if p == nil {
				err = fmt.Errorf("project %q not found", args[0])
				return common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			args[0] = p.ToProto().ProjectId
		}

		id, err := sdktypes.ParseProjectID(args[0])
		if err != nil {
			return fmt.Errorf("invalid project ID %q: %w", args[0], err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		bs, err := builds().List(ctx, sdkservices.ListBuildsFilter{ProjectID: id})
		if err != nil {
			return fmt.Errorf("list builds: %w", err)
		}

		if len(bs) == 0 {
			return common.FailNotFound(cmd, "builds")
		}

		common.RenderList(bs)
		return nil
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(listCmd)
}
