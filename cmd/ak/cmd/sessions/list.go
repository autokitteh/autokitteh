package sessions

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	stateType     stateString
	withInputs    bool
	nextPageToken string
	pageSize      int
	skipRows      int
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [filter flags] [--fail] [--with-inputs]",
	Short:   "List all sessions",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		f := sdkservices.ListSessionsFilter{}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		if deploymentID != "" {
			d, did, err := r.DeploymentID(ctx, deploymentID)
			if err = common.AddNotFoundErrIfCond(err, d.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "deployment")
			}
			f.DeploymentID = did
		}

		if project != "" {
			pid, err := r.ProjectNameOrID(ctx, sdktypes.InvalidOrgID, project)
			if err = common.AddNotFoundErrIfCond(err, pid.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "project")
			}
			f.ProjectID = pid
		}

		if eventID != "" {
			e, eid, err := r.EventID(ctx, eventID)
			if err = common.AddNotFoundErrIfCond(err, e.IsValid()); err != nil {
				return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "event")
			}
			f.EventID = eid
		}

		if nextPageToken != "" {
			f.PageToken = nextPageToken
		}

		if pageSize > 0 {
			f.PageSize = int32(pageSize)
		}

		if skipRows > 0 {
			f.Skip = int32(skipRows)
		}

		var err error
		if f.StateType, err = sdktypes.ParseSessionStateType(stateType.String()); err != nil {
			return fmt.Errorf("invalid state %q: %w", stateType, err)
		}

		result, err := sessions().List(ctx, f)
		if result == nil {
			result = &sdkservices.ListSessionResult{}
		}
		err = common.AddNotFoundErrIfCond(err, len(result.Sessions) > 0)
		if err = common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "sessions"); err == nil {
			if !withInputs {
				for i := range result.Sessions {
					result.Sessions[i] = result.Sessions[i].WithInputs(nil)
				}
			}
			common.RenderList(result.Sessions)
			if result.NextPageToken != "" {
				common.RenderKV("next-page-token", result.NextPageToken)
			}
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	listCmd.Flags().StringVarP(&deploymentID, "deployment-id", "d", "", "deployment ID")
	listCmd.Flags().StringVar(&eventID, "event-id", "", "event ID")
	listCmd.Flags().VarP(&stateType, "state-type", "s", strings.Join(possibleStates, "|"))
	listCmd.Flags().BoolVarP(&withInputs, "with-inputs", "i", false, "include input details")
	listCmd.Flags().StringVar(&nextPageToken, "next-page-token", "", "provide the returned page token to get next")
	listCmd.Flags().IntVar(&pageSize, "page-size", 50, "page size")
	listCmd.Flags().IntVar(&skipRows, "skip-rows", 0, "skip rows")

	common.AddFailIfNotFoundFlag(listCmd)
}
