package connections

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	project string
	org     string
	quiet   bool
)

var createCmd = common.StandardCommand(&cobra.Command{
	Use:     "create <name> <--project=...> <--integration=...> [--quiet]",
	Short:   "Define new connection to integration",
	Aliases: []string{"c"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		cli := common.Client()
		r := resolver.Resolver{Client: cli}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		if org == "" {
			u, err := cli.Auth().WhoAmI(ctx)
			if err != nil {
				err := fmt.Errorf("org not provided and could not resolve current user: %w", err)
				return common.NewExitCodeError(common.UnauthenticatedExitCode, err)
			}

			org = u.DefaultOrgID().String()
		}

		org, err := r.Org(ctx, org)
		if err != nil {
			err := fmt.Errorf("resolve org: %w", err)
			return common.NewExitCodeError(common.GenericFailureExitCode, err)
		}

		orgID := org.String()

		scope := sdktypes.ConnectionScopeOrg
		pidStr := ""

		if project != "" {
			pid, err := r.ProjectNameOrID(ctx, sdktypes.InvalidOrgID, project)
			if err != nil {
				return err
			}
			if !pid.IsValid() {
				err = fmt.Errorf("project %q not found", project)
				return common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			scope = sdktypes.ConnectionScopeProject
			pidStr = pid.String()
		}

		i, iid, err := r.IntegrationNameOrID(ctx, integration)
		if err != nil {
			return err
		}
		if !i.IsValid() {
			err = fmt.Errorf("integration %q not found", integration)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		c, err := sdktypes.ConnectionFromProto(&sdktypes.ConnectionPB{
			IntegrationId: iid.String(),
			ProjectId:     pidStr,
			OrgId:         orgID,
			Name:          args[0],
			Scope:         scope,
		})
		if err != nil {
			return fmt.Errorf("invalid connection: %w", err)
		}

		cid, err := connections().Create(ctx, c)
		if err != nil {
			return fmt.Errorf("create connection: %w", err)
		}

		if !quiet {
			conn, err := connections().Get(ctx, cid)
			if err != nil {
				return fmt.Errorf("get connection: %w", err)
			}

			if l := conn.Links().InitURL(); l != "" {
				action := "and can be initialized"
				if conn.Capabilities().RequiresConnectionInit() {
					action = "but requires initialization"
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "Connection created, %s. Please run this to complete: ak connection init %v\n", action, cid)
			}
		}

		common.RenderKV("connection_id", cid)
		return nil
	},
})

func init() {
	// Command-specific flags.
	createCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID (creates a project-scoped connection)")
	createCmd.Flags().StringVarP(&org, "org", "o", "", "organization name or ID (creates an org-scoped connection)")
	// createCmd.MarkFlagsOneRequired("project", "org")
	createCmd.MarkFlagsMutuallyExclusive("project", "org")

	createCmd.Flags().StringVarP(&integration, "integration", "i", "", "integration name or ID")
	kittehs.Must0(createCmd.MarkFlagRequired("integration"))

	createCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "do not print initialization guidance")
}
