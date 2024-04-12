package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/builds"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/configuration"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/connections"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/deployments"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/envs"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/events"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/experimental"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/integrations"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/manifest"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/projects"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/runtimes"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/server"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/sessions"
	"go.autokitteh.dev/autokitteh/cmd/ak/cmd/triggers"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/config"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
)

var (
	configs []string

	debugLogs, json, niceJSON bool
)

var RootCmd = common.StandardCommand(&cobra.Command{
	Use:   "ak",
	Short: "autokitteh command-line interface and local server",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Set the output renderer based on global flags.
		if json {
			common.SetRenderer(common.JSONRenderer)
		}
		if niceJSON {
			common.SetRenderer(common.NiceJSONRenderer)
		}

		// Initialize all the configurations.
		path := filepath.Join(xdg.ConfigHomeDir(), ".env")
		if err := godotenv.Load(path); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf(".env loading error: %w", err)
			}
		}

		confmap, err := parseConfigs(configs)
		if err != nil {
			return err
		}

		if debugLogs {
			confmap["logger.zap.level"] = "debug"
		}

		if err := common.InitConfig(confmap); err != nil {
			return fmt.Errorf("root init config: %w", err)
		}
		cfg := common.Config()

		url := sdkclient.DefaultLocalURL
		if _, err := cfg.Get(config.ServiceUrlConfigKey, &url); err != nil {
			return fmt.Errorf("failed parse config: %w", err)
		} // if not overriden by config, then url will remain default
		if !strings.HasPrefix(url, "http") {
			url = "http://" + url
		}

		common.InitRPCClient(url, "")

		return nil
	},
})

// Execute is the central point to run any command with corresponding
// parameters and flags, and to handle errors in a standard way.
// This is called only once by [main.main].
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		var ecerr common.ExitCodeError
		if errors.As(err, &ecerr) {
			os.Exit(ecerr.Code)
		}
		os.Exit(1)
	}
}

func init() {
	// Global flags for all commands.
	RootCmd.PersistentFlags().StringArrayVarP(&configs, "config", "c", nil, `temporary "key=value" configurations`)
	RootCmd.PersistentFlags().BoolVar(&debugLogs, "debug", false, `emit debug logs`)

	RootCmd.PersistentFlags().BoolVarP(&json, "json", "j", false, "print output in compact JSON format")
	RootCmd.PersistentFlags().BoolVarP(&niceJSON, "nice_json", "J", false, "print output in readable JSON format")
	RootCmd.MarkFlagsMutuallyExclusive("json", "nice_json")

	// Top-level standalone commands.
	RootCmd.AddCommand(completionCmd)
	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(upCmd)
	RootCmd.AddCommand(versionCmd)

	// Top-level parent commands.
	builds.AddSubcommands(RootCmd)
	configuration.AddSubcommands(RootCmd)
	connections.AddSubcommands(RootCmd)
	deployments.AddSubcommands(RootCmd)
	envs.AddSubcommands(RootCmd)
	events.AddSubcommands(RootCmd)
	experimental.AddSubcommands(RootCmd)
	integrations.AddSubcommands(RootCmd)
	manifest.AddSubcommands(RootCmd)
	projects.AddSubcommands(RootCmd)
	runtimes.AddSubcommands(RootCmd)
	server.AddSubcommands(RootCmd)
	sessions.AddSubcommands(RootCmd)
	triggers.AddSubcommands(RootCmd)
}

func parseConfigs(pairs []string) (map[string]any, error) {
	confmap := make(map[string]any, len(pairs))
	for _, s := range pairs {
		k, v, ok := strings.Cut(s, "=")
		if !ok {
			return nil, fmt.Errorf(`invalid config argument %q, expected "key=value"`, s)
		}

		confmap[k] = v
	}
	return confmap, nil
}
