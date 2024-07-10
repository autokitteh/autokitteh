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
			if err != nil {
				return err
			}
			if !d.IsValid() {
				err = fmt.Errorf("deployment ID %q not found", deploymentID)
				return common.NewExitCodeError(common.NotFoundExitCode, err)
			}
			f.DeploymentID = did
		}

		if env != "" {
			e, _, err := r.EnvNameOrID(ctx, env, "")
			if err != nil {
				return err
			}
			f.EnvID = e.ID()
		}

		if eventID != "" {
			e, eid, err := r.EventID(ctx, eventID)
			if err != nil {
				return err
			}
			if !e.IsValid() {
				err = fmt.Errorf("event ID %q not found", eventID)
				return common.NewExitCodeError(common.NotFoundExitCode, err)
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
		if err != nil {
			return fmt.Errorf("list sessions: %w", err)
		}

		if err := common.FailIfNotFound(cmd, "sessions", len(result.Sessions) > 0); err != nil {
			return err
		}

		if !withInputs {
			for i := range result.Sessions {
				result.Sessions[i] = result.Sessions[i].WithInputs(nil)
			}
		}

		common.RenderList(result.Sessions)

		if result.NextPageToken != "" {
			common.RenderKV("next-page-token", result.NextPageToken)
		}
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	listCmd.Flags().StringVarP(&deploymentID, "deployment-id", "d", "", "deployment ID")
	listCmd.Flags().StringVar(&eventID, "event-id", "", "event ID")
	listCmd.Flags().VarP(&stateType, "state-type", "s", strings.Join(possibleStates, "|"))
	listCmd.Flags().BoolVarP(&withInputs, "with-inputs", "i", false, "include input details")
	listCmd.Flags().StringVar(&nextPageToken, "next-page-token", "", "provide the returned page token to get next")
	listCmd.Flags().IntVar(&pageSize, "page-size", 50, "page size")
	listCmd.Flags().IntVar(&skipRows, "skip-rows", 0, "skip rows")

	common.AddFailIfNotFoundFlag(listCmd)
}
