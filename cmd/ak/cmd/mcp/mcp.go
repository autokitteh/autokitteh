package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/version"
)

var (
	stdio, sse          bool
	sseAddr, sseBaseURL string
)

var mcpCmd = common.StandardCommand(&cobra.Command{
	Use:   "mcp {--stdio|--sse [--sse-addr <addr>] [--sse-base-url <base-url>]}",
	Short: "Start local MCP server",
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		srv := server.NewMCPServer(
			"AutoKitteh",
			version.Version,
			server.WithResourceCapabilities(true, true),
			server.WithLogging(),
		)

		srv.AddTools(tools...)
		addResources(srv)

		if stdio {
			if err := server.ServeStdio(srv); err != nil {
				return err
			}
		} else {
			sseSrv := server.NewSSEServer(
				srv,
				server.WithBaseURL(sseBaseURL),
			)

			fmt.Fprintf(cmd.OutOrStdout(), "starting SSE server on %s, url=%s.\n", sseAddr, sseSrv.CompleteSseEndpoint())

			if err := sseSrv.Start(sseAddr); err != nil {
				return err
			}
		}

		return nil
	},
})

func init() {
	mcpCmd.Flags().BoolVar(&stdio, "stdio", false, "stdio server")
	mcpCmd.Flags().BoolVar(&sse, "sse", false, "sse server")
	mcpCmd.Flags().StringVarP(&sseBaseURL, "sse-base-url", "u", "http://localhost:3000", "sse server base url")
	mcpCmd.Flags().StringVarP(&sseAddr, "sse-addr", "a", "localhost:3000", "sse server addr")
	mcpCmd.MarkFlagsMutuallyExclusive("stdio", "sse")
	mcpCmd.MarkFlagsOneRequired("stdio", "sse")
}

func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(mcpCmd)
}
