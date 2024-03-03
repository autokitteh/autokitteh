package connections

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var project string

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create <name> <--project=...> <--integration=...> <--connection-token=...>",
	Short:   "Define new connection to integration",
	Aliases: []string{"c"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		p, pid, err := r.ProjectNameOrID(project)
		if err != nil {
			return err
		}
		if !p.IsValid() {
			err = fmt.Errorf("project %q not found", project)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		i, iid, err := r.IntegrationNameOrID(integration)
		if err != nil {
			return err
		}
		if !i.IsValid() {
			err = fmt.Errorf("integration %q not found", integration)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		c, err := sdktypes.ConnectionFromProto(&sdktypes.ConnectionPB{
			IntegrationId:    iid.String(),
			IntegrationToken: connectionToken,
			ProjectId:        pid.String(),
			Name:             args[0],
		})
		if err != nil {
			return fmt.Errorf("invalid connection: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		cid, err := connections().Create(ctx, c)
		if err != nil {
			return fmt.Errorf("create connection: %w", err)
		}

		common.RenderKV("connection_id", cid)
		return nil
	},
})

func init() {
	// Command-specific flags.
	createCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("project"))

	createCmd.Flags().StringVarP(&integration, "integration", "i", "", "integration name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("integration"))

	createCmd.Flags().StringVarP(&connectionToken, "connection-token", "t", "", "connection token")
	kittehs.Must0(createCmd.MarkFlagRequired("connection-token"))
}
