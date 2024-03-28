package sessions

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Default flag value shared by the "start", "restart", and "watch" subcommands.
const (
	defaultPollInterval = 1 * time.Second
)

var (
	// Flags shared by the "list" and "start" subcommands.
	deploymentID, env, eventID string

	// Flags shared by the "start", "restart", and "watch" subcommands.
	pollInterval time.Duration
	watchTimeout time.Duration
	watch, quiet bool
	noTimestamps bool
)

var sessionsCmd = common.StandardCommand(&cobra.Command{
	Use:     "sessions",
	Short:   "Session management commands",
	Aliases: []string{"sessions", "session", "sess"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(sessionsCmd)
}

func init() {
	// Subcommands.
	sessionsCmd.AddCommand(deleteCmd)
	sessionsCmd.AddCommand(getCmd)
	sessionsCmd.AddCommand(listCmd)
	sessionsCmd.AddCommand(logCmd)
	sessionsCmd.AddCommand(restartCmd)
	sessionsCmd.AddCommand(startCmd)
	sessionsCmd.AddCommand(stopCmd)
	sessionsCmd.AddCommand(testCmd)
	sessionsCmd.AddCommand(watchCmd)
}

func sessions() sdkservices.Sessions {
	return common.Client().Sessions()
}

func latestSessionID() (string, error) {
	ctx, cancel := common.LimitedContext()
	defer cancel()

	ss, _, err := sessions().List(ctx, sdkservices.ListSessionsFilter{})
	if err != nil {
		return "", fmt.Errorf("list sessions: %w", err)
	}

	if len(ss) == 0 {
		return "", common.NewExitCodeError(common.NotFoundExitCode, errors.New("sessions not found"))
	}

	latest := ss[0]
	for _, s := range ss[1:] {
		if s.ToProto().CreatedAt.AsTime().After(latest.ToProto().CreatedAt.AsTime()) {
			latest = s
		}
	}
	return string(latest.ToProto().SessionId), nil
}
