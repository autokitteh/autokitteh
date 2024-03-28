package sessions

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	buildID    string
	entryPoint string
	memos      []string
)

var startCmd = common.StandardCommand(&cobra.Command{
	Use:   "start {--deployment-id <ID>|--build-id <ID> --env <name or ID>} --entrypoint <...> [--memo <...>] [--watch [--watch-timeout <duration>] [--poll-interval <duration>] [--no-timestamps] [--quiet]]",
	Short: "Start new session",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		did, eid, bid, ep, err := sessionArgs()
		if err != nil {
			return err
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		s := sdktypes.NewSession(bid, ep, nil, nil).WithEnvID(eid).WithDeploymentID(did)
		sid, err := sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}

		common.RenderKVIfV("session_id", sid)

		if watch {
			_, err := sessionWatch(sid, sdktypes.SessionStateTypeUnspecified)
			return err
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	startCmd.Flags().StringVarP(&deploymentID, "deployment-id", "d", "", "deployment ID, mutually exclusive with --build-id and --env")
	startCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "build ID, mutually exclusive with --deployment-id")
	startCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID, mutually exclusive with --deployment-id")
	startCmd.MarkFlagsOneRequired("deployment-id", "build-id")
	startCmd.MarkFlagsRequiredTogether("build-id", "env")

	startCmd.Flags().StringVarP(&entryPoint, "entrypoint", "p", "", `entry point ("file:function")`)
	kittehs.Must0(startCmd.MarkFlagRequired("entrypoint"))

	startCmd.Flags().StringSliceVarP(&memos, "memo", "m", nil, `zero or more "key=value" pairs`)

	startCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch session to completion")

	startCmd.Flags().DurationVarP(&watchTimeout, "watch-timeout", "t", 0, "watch timeout duration")
	startCmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", defaultPollInterval, "watch poll interval")
	startCmd.Flags().BoolVarP(&noTimestamps, "no-timestamps", "n", false, "omit timestamps from watch output")
	startCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "don't print anything, just wait to finish")
}

func sessionArgs() (did sdktypes.DeploymentID, eid sdktypes.EnvID, bid sdktypes.BuildID, ep sdktypes.CodeLocation, err error) {
	r := resolver.Resolver{Client: common.Client()}

	if deploymentID != "" {
		var d sdktypes.Deployment
		if d, did, err = r.DeploymentID(deploymentID); err != nil {
			return
		}
		if !d.IsValid() {
			err = fmt.Errorf("deployment %q not found", deploymentID)
			err = common.NewExitCodeError(common.NotFoundExitCode, err)
			return
		}

		bid, eid = d.BuildID(), d.EnvID()
	}

	if buildID != "" {
		var b sdktypes.Build
		if b, bid, err = r.BuildID(buildID); err != nil {
			return
		}
		if !b.IsValid() {
			err = fmt.Errorf("build %q not found", buildID)
			err = common.NewExitCodeError(common.NotFoundExitCode, err)
			return
		}

	}

	if env != "" {
		var e sdktypes.Env
		if e, eid, err = r.EnvNameOrID(env, ""); err != nil {
			return
		}
		if env != "" && !e.IsValid() {
			err = fmt.Errorf("env %q not found", eventID)
			err = common.NewExitCodeError(common.NotFoundExitCode, err)
			return
		}
	}

	if ep, err = sdktypes.ParseCodeLocation(entryPoint); err != nil {
		err = fmt.Errorf("invalid entry-point %q: %w", entryPoint, err)
		return
	}

	return
}
