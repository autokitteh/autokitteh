package runtimes

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/configrt"
	"go.autokitteh.dev/autokitteh/runtimes/pythonrt"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flag shared by all subcommands.
var local bool

var runtimeCmd = common.StandardCommand(&cobra.Command{
	Use:     "runtime",
	Short:   "Runtime engines: build, get, list, run",
	Aliases: []string{"rt"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(runtimeCmd)
}

func init() {
	// Flag shared by all subcommands.
	runtimeCmd.PersistentFlags().BoolVarP(&local, "local", "l", false, "execute locally")

	// Subcommands.
	runtimeCmd.AddCommand(buildCmd)
	runtimeCmd.AddCommand(getCmd)
	runtimeCmd.AddCommand(listCmd)
	runtimeCmd.AddCommand(runCmd)
	runtimeCmd.AddCommand(testCmd)
}

func runtimes() sdkservices.Runtimes {
	if local {
		return kittehs.Must1(sdkruntimes.New([]*sdkruntimes.Runtime{
			starlarkrt.New(),
			configrt.New(),
			kittehs.Must1(pythonrt.New(
				&pythonrt.Config{LazyLoadLocalVEnv: true},
				zap.NewNop(),
				func() string { return "localhost" },
			)),
		}))
	}
	return common.Client().Runtimes()
}
