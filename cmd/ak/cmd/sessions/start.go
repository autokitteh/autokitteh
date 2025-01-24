package sessions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	buildID    string
	entryPoint string
	memos      []string
	inputs     []string
)

var startCmd = common.StandardCommand(&cobra.Command{
	Use:   "start {--deployment-id <ID>|--build-id <ID> --project <name or ID>} --entrypoint <...> [--memo <...>] [--input <JSON> [...]] [--watch [--watch-timeout <duration>] [--poll-interval <duration>] [--no-timestamps] [--quiet]]",
	Short: "Start new session",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		did, pid, bid, ep, err := sessionArgs()
		if err != nil {
			return err
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		inputs, err := parseinputs()
		if err != nil {
			return err
		}

		s := sdktypes.NewSession(bid, ep, nil, nil).WithProjectID(pid).WithDeploymentID(did).WithInputs(inputs)
		sid, err := sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}

		common.RenderKVIfV("session_id", sid)

		if watch {
			_, err := sessionWatch(sid, sdktypes.SessionStateTypeUnspecified, "")
			return err
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	startCmd.Flags().StringVarP(&deploymentID, "deployment-id", "d", "", "deployment ID, mutually exclusive with --build-id and --env")
	startCmd.Flags().StringVarP(&buildID, "build-id", "b", "", "build ID, mutually exclusive with --deployment-id")
	startCmd.Flags().StringVarP(&project, "project", "o", "", "project name or ID, mutually exclusive with --deployment-id")
	startCmd.MarkFlagsOneRequired("deployment-id", "build-id")

	startCmd.Flags().StringVarP(&entryPoint, "entrypoint", "p", "", `entry point ("file:function")`)

	startCmd.Flags().StringSliceVarP(&memos, "memo", "m", nil, `zero or more "key=value" pairs`)

	startCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch session to completion")

	startCmd.Flags().DurationVarP(&watchTimeout, "watch-timeout", "t", 0, "watch timeout duration")
	startCmd.Flags().DurationVarP(&pollInterval, "poll-interval", "i", defaultPollInterval, "watch poll interval")
	startCmd.Flags().BoolVarP(&noTimestamps, "no-timestamps", "n", false, "omit timestamps from watch output")
	startCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "don't print anything, just wait to finish")

	startCmd.Flags().StringArrayVarP(&inputs, "input", "I", nil, `zero or more "key=value" pairs, where value is a JSON value`)
}

func parseinputs() (map[string]sdktypes.Value, error) {
	m := make(map[string]sdktypes.Value, len(inputs))
	for _, v := range inputs {
		k, v, ok := strings.Cut(v, "=")
		if !ok {
			return nil, fmt.Errorf("invalid value %q", v)
		}

		decoder := json.NewDecoder(strings.NewReader(v))
		decoder.UseNumber()

		var jv any
		if err := decoder.Decode(&jv); err != nil {
			return nil, fmt.Errorf("invalid value %q: %w", v, err)
		}

		wv, err := sdktypes.WrapValue(jv)
		if err != nil {
			return nil, fmt.Errorf("unhandled value type for %q: %w", v, err)
		}

		m[k] = wv
	}

	return m, nil
}

func sessionArgs() (did sdktypes.DeploymentID, pid sdktypes.ProjectID, bid sdktypes.BuildID, ep sdktypes.CodeLocation, err error) {
	r := resolver.Resolver{Client: common.Client()}
	ctx, cancel := common.LimitedContext()
	defer cancel()

	if deploymentID != "" {
		var d sdktypes.Deployment
		if d, did, err = r.DeploymentID(ctx, deploymentID); err != nil {
			return
		}
		if !d.IsValid() {
			err = fmt.Errorf("deployment %q not found", deploymentID)
			err = common.NewExitCodeError(common.NotFoundExitCode, err)
			return
		}

		bid, pid = d.BuildID(), d.ProjectID()
	}

	if buildID != "" {
		var b sdktypes.Build
		if b, bid, err = r.BuildID(ctx, buildID); err != nil {
			return
		}
		if !b.IsValid() {
			err = fmt.Errorf("build %q not found", buildID)
			err = common.NewExitCodeError(common.NotFoundExitCode, err)
			return
		}

	}

	if project != "" {
		if pid, err = r.ProjectNameOrID(ctx, sdktypes.InvalidOrgID, project); err != nil {
			return
		}
		if project != "" && !pid.IsValid() {
			err = fmt.Errorf("project %q not found", project)
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
