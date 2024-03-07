package sessions

import (
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
	buildID    string
	memos      []string
)

var startCmd = common.StandardCommand(&cobra.Command{
	Use:   "start <--deployment-id=... | --build-id=...> <--event-id=...> <--entrypoint=...> [--memo=...] [--watch] [--poll-interval=DURATION] [--watch-timeout=DURATION] [--no-timestamps]",
	Short: "Start new session",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		if buildID != "" && deploymentID != "" {
			return fmt.Errorf("--deployment-id and --build-id are mutually exclusive")
		}

		r := resolver.Resolver{Client: common.Client()}

		var (
			did sdktypes.DeploymentID
			bid sdktypes.BuildID
		)

		if deploymentID != "" {
			var err error
			if _, did, err = r.DeploymentID(deploymentID); err != nil {
				return fmt.Errorf("deployment ID %q: %w", deploymentID, err)
			}
		}

		if buildID != "" {
			var err error
			if bid, err = sdktypes.ParseBuildID(buildID); err != nil {
				return fmt.Errorf("build ID %q: %w", buildID, err)
			}
		}

		var e sdktypes.Event
		if eventID != "" {
			var err error

			e, _, err = r.EventID(eventID)
			if err != nil {
				return err
			}
			if !e.IsValid() {
				err = fmt.Errorf("event ID %q not found", eventID)
				return common.NewExitCodeError(common.NotFoundExitCode, err)
			}
		}

		ep, err := sdktypes.StrictParseCodeLocation(entryPoint)
		if err != nil {
			return fmt.Errorf("invalid entry-point %q: %w", entryPoint, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		s := sdktypes.NewSession(did, bid, sdktypes.InvalidSessionID, e.ID(), ep, e.ToValues(), nil)
		sid, err := sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}

		common.RenderKVIfV("session_id", sid)

		if track {
			return sessionWatch(sid, sdktypes.SessionStateTypeUnspecified)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	startCmd.Flags().StringVarP(&deploymentID, "deployment-id", "d", "", "deployment ID")
	startCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "build ID")
	startCmd.Flags().StringVarP(&eventID, "event-id", "e", "", "event ID")
	startCmd.Flags().StringVarP(&entryPoint, "entrypoint", "p", "", `entry point ("file:function")`)
	kittehs.Must0(startCmd.MarkFlagRequired("entrypoint"))
	startCmd.Flags().StringSliceVarP(&memos, "memo", "m", nil, `zero or more "key=value" pairs`)
	startCmd.Flags().BoolVarP(&track, "watch", "w", false, "watch session to completion")
	startCmd.Flags().DurationVar(&pollInterval, "poll-interval", defaultPollInterval, "poll interval")
	startCmd.Flags().BoolVar(&noTimestamps, "no-timestamps", false, "omit timestamps from track output")
	startCmd.Flags().DurationVar(&watchTimeout, "watch-timeout", 0, "watch time out duration")
}
