package sessions

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	defaultWait = 1 * time.Second
)

var (
	entryPoint   string
	memos        []string
	wait         bool
	waitInterval time.Duration
)

var startCmd = common.StandardCommand(&cobra.Command{
	Use:   "start <--deployment-id=...> <--event-id=...> <--entrypoint=...> [--memo=...] [--wait]",
	Short: "Start new session",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		d, did, err := r.DeploymentID(deploymentID)
		if err != nil {
			return err
		}
		if !d.IsValid() {
			err = fmt.Errorf("deployment ID %q not found", deploymentID)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		e, eid, err := r.EventID(eventID)
		if err != nil {
			return err
		}
		if !e.IsValid() {
			err = fmt.Errorf("event ID %q not found", eventID)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		ep, err := sdktypes.StrictParseCodeLocation(entryPoint)
		if err != nil {
			return fmt.Errorf("invalid entry-point %q: %w", entryPoint, err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		s := sdktypes.NewSession(did, sdktypes.InvalidSessionID, eid, ep, e.ToValues(), nil)
		sid, err := sessions().Start(ctx, s)
		if err != nil {
			return fmt.Errorf("start session: %w", err)
		}

		common.RenderKVIfV("session_id", sid)

		if wait {
			fmt.Println()
			e := make(chan error)
			done := make(chan bool)
			state := make(chan string)
			go waitForSession(sid, e, done, state)
			go updateStateTicker(e, done, state)
			select {
			case err := <-e:
				return err
			case <-done:
				break
			}
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	startCmd.Flags().StringVarP(&deploymentID, "deployment-id", "d", "", "deployment ID")
	kittehs.Must0(startCmd.MarkFlagRequired("deployment-id"))

	startCmd.Flags().StringVarP(&eventID, "event-id", "e", "", "event ID")
	kittehs.Must0(startCmd.MarkFlagRequired("event-id"))

	startCmd.Flags().StringVarP(&entryPoint, "entrypoint", "p", "", `entry point ("file:function")`)
	kittehs.Must0(startCmd.MarkFlagRequired("entrypoint"))

	startCmd.Flags().StringSliceVarP(&memos, "memo", "m", nil, `zero or more "key=value" pairs`)
	startCmd.Flags().BoolVarP(&wait, "wait", "w", false, "wait for session to complete")
	startCmd.Flags().DurationVarP(&waitInterval, "wait-interval", "i", defaultWait, "wait interval")
}

// waitForSession runs as a goroutine and waits for the session state to be
// either ERROR or COMPLETED. In the meantime it prints the current state
// with a time ticker. Session retrieval errors and session failures are
// both reported as errors by the CLI.
func waitForSession(id sdktypes.SessionID, e chan<- error, done chan<- bool, state chan<- string) {
	errorState := sdktypes.SessionStateTypeError.ToProto()
	completedState := sdktypes.SessionStateTypeCompleted.ToProto()
	startTime := time.Now()

	for {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		s, err := sessions().Get(ctx, id)
		if err != nil {
			e <- fmt.Errorf("get session state: %w", err)
			return
		}
		ss := s.ToProto().State
		state <- ss.String()

		if ss == errorState {
			e <- errors.New("session failed")
			return
		}
		if ss == completedState {
			done <- true
			printStateTicker(startTime, ss.String())
			return
		}

		time.Sleep(waitInterval)
	}
}

// updateStateTicker runs as a goroutine and updates a time ticker
// with the current session state once every second.
func updateStateTicker(e <-chan error, done <-chan bool, state <-chan string) {
	startTime := time.Now()
	currentState := "UNSPECIFIED"

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e:
			return
		case <-done:
			return
		case currentState = <-state:
			continue
		case <-ticker.C:
			printStateTicker(startTime, currentState)
		}
	}
}

func printStateTicker(startTime time.Time, currentState string) {
	duration := time.Since(startTime).Round(time.Second)
	cs := strings.ReplaceAll(currentState, "SESSION_STATE_TYPE_", "")
	fmt.Printf("\033[Fduration %s - current state %s\n", duration, cs)
}
