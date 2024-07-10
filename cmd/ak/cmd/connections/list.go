package connections

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--integration=...] [--project=...] [--fail]",
	Short:   "List all connections",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		var f sdkservices.ListConnectionsFilter

		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		if integration != "" {
			_, iid, err := r.IntegrationNameOrID(ctx, integration)
			if err != nil {
				return err
			}

			if !iid.IsValid() {
				return fmt.Errorf("integration %q not found", integration)
			}
			f.IntegrationID = iid
		}

		if project != "" {
			_, pid, err := r.ProjectNameOrID(ctx, project)
			if err != nil {
				return err
			}

			if !pid.IsValid() {
				return fmt.Errorf("project %q not found", integration)
			}
			f.ProjectID = pid
		}

		cs, err := connections().List(ctx, f)
		if err != nil {
			return fmt.Errorf("list connections: %w", err)
		}

		if err := common.FailIfNotFound(cmd, "connections", len(cs) > 0); err != nil {
			return err
		}

		common.RenderList(cs)
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&integration, "integration", "i", "", "integration name or ID")
	listCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")

	common.AddFailIfNotFoundFlag(listCmd)
}
