package triggers

import (
	"errors"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [-p project] [-e env] [-c conection] [--fail]",
	Short:   "List all event triggers",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}

		// All flags are optional.
		p, pid, err := r.ProjectNameOrID(project)
		if err != nil {
			return err
		}
		if project != "" && !p.IsValid() {
			err = resolver.NotFoundError{Type: "project", Name: project}
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		e, eid, err := r.EnvNameOrID(env, project)
		if err != nil {
			return err
		}
		if env != "" && !e.IsValid() {
			err = resolver.NotFoundError{Type: "environment", Name: env}
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		c, cid, err := r.ConnectionNameOrID(connection, project)
		if err != nil {
			if errors.As(err, resolver.NotFoundErrorType) {
				err = common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			return err
		}
		if cid.IsValid() && !c.IsValid() {
			err = resolver.NotFoundError{Type: "connection ID", Name: connection}
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		ts, err := triggers().List(ctx, sdkservices.ListTriggersFilter{
			ProjectID:    pid,
			EnvID:        eid,
			ConnectionID: cid,
		})
		if err != nil {
			return err
		}

		if err := common.FailIfNotFound(cmd, "triggers", len(ts) > 0); err != nil {
			return err
		}

		common.RenderList(ts)
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	listCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	listCmd.Flags().StringVarP(&connection, "connection", "c", "", "connection name or ID")

	common.AddFailIfNotFoundFlag(listCmd)
}
