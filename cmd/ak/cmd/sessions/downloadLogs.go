package sessions

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
)

var outputPath string

var downloadLogsCmd = common.StandardCommand(&cobra.Command{
	Use:   "download-logs [session ID]",
	Short: "Download logs for a session",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		sid, err := acquireSessionID(args[0])
		if err = common.AddNotFoundErrIfCond(err, sid.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		ctx, done := common.LimitedContext()
		defer done()

		data, err := sessions().DownloadLogs(ctx, sid)
		if err != nil {
			return fmt.Errorf("failed to download logs: %w", err)
		}

		// Use default output filename if none provided.
		if outputPath == "" {
			timestamp := time.Now().Format("20060102_150405")
			filename := fmt.Sprintf("%s_%s.txt", sid.String(), timestamp)
			outputPath = filepath.Join(".", filename)
		}

		if err := os.WriteFile(outputPath, data, 0o644); err != nil {
			return fmt.Errorf("failed to write to file %q: %w", outputPath, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Logs written to %s\n", outputPath)
		return nil
	},
})

func init() {
	downloadLogsCmd.Flags().StringVarP(&outputPath, "output", "o", "", "path to output file (default is ./<session_id>_<timestamp>.txt)")
	common.AddFailIfNotFoundFlag(downloadLogsCmd)
}
