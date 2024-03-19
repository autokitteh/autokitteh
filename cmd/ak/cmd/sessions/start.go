package sessions

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	defaultPollInterval = 1 * time.Second
)

var (
	entryPoint string
	memos      []string
	build      string
	deployID   string
)

var startCmd = common.StandardCommand(&cobra.Command{
	Use:   "start [--build-id=...] [--env=...] [--deployment-id=...] <--entrypoint=...> [--memo=...] [--watch] [--watch-timeout=...] [--poll-interval=...] [--no-timestamps] [--quiet]",
	Short: "Start new session",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		if deployID == "" && build == "" {
			return errors.New("either --deployment-id or --build-id must be provided")
		}

		if deployID != "" && (build != "" || env != "") {
			return errors.New("--deployment-id cannot be used with --build-id or --env")
		}

		did, eid, bid, ep, err := sessionArgs()
		if err != nil {
			return err
		}

		if !ep.IsValid() {
			return errors.New("--entrypoint must be specified")
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		s := sdktypes.NewSession(bid, ep, nil, nil).WithEnvID(eid).WithDeploymentID(did)
		sid, err := sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}

		common.RenderKVIfV("session_id", sid)

		if track {
			_, err := sessionWatch(sid, sdktypes.SessionStateTypeUnspecified)
			return err
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	startCmd.Flags().StringVarP(&deployID, "deployment-id", "d", "", "deployment ID, mutually exclusive with --build-id and --env")
	startCmd.Flags().StringVarP(&build, "build-id", "b", "", "build ID")
	startCmd.Flags().StringVar(&env, "env", "", "env")

	startCmd.Flags().StringVarP(&entryPoint, "entrypoint", "p", "", `entry point ("file:function")`)
	kittehs.Must0(startCmd.MarkFlagRequired("entrypoint"))

	startCmd.Flags().StringSliceVarP(&memos, "memo", "m", nil, `zero or more "key=value" pairs`)
	startCmd.Flags().BoolVarP(&track, "watch", "w", false, "watch session to completion")
	startCmd.Flags().DurationVar(&pollInterval, "poll-interval", defaultPollInterval, "poll interval")

	startCmd.Flags().BoolVar(&noTimestamps, "no-timestamps", false, "omit timestamps from track output")
	startCmd.Flags().DurationVar(&watchTimeout, "watch-timeout", 0, "watch time out duration")
	startCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "do not print anything, just wait to finish")
}

func sessionArgs() (did sdktypes.DeploymentID, eid sdktypes.EnvID, bid sdktypes.BuildID, ep sdktypes.CodeLocation, err error) {
	r := resolver.Resolver{Client: common.Client()}

	if deployID != "" {
		var d sdktypes.Deployment
		if d, did, err = r.DeploymentID(deployID); err != nil {
			return
		}
		if !d.IsValid() {
			err = fmt.Errorf("deployment %q not found", deployID)
			err = common.NewExitCodeError(common.NotFoundExitCode, err)
			return
		}

		bid, eid = d.BuildID(), d.EnvID()
	}

	if build != "" {
		var b sdktypes.Build
		if b, bid, err = r.BuildID(build); err != nil {
			return
		}
		if !b.IsValid() {
			err = fmt.Errorf("build %q not found", build)
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
