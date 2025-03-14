package sessions

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var printsCmd = common.StandardCommand(&cobra.Command{
	Use:   "prints [sessions ID | project] [--fail] [--no-timestamps]",
	Short: "Get session prints",
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		sid, err := acquireSessionID(args[0])
		if err = common.AddNotFoundErrIfCond(err, sid.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "session")
		}

		ctx, done := common.LimitedContext()
		defer done()

		prints, err := sessions().GetPrints(ctx, sid, sdktypes.PaginationRequest{})
		if err != nil {
			return fmt.Errorf("get log: %w", err)
		}

		for _, p := range prints.Prints {
			text, err := p.Value.ToString()
			if err != nil {
				text = fmt.Sprintf("error converting print to string: %v", err.Error())
			}

			if !noTimestamps {
				fmt.Printf("[%s] ", p.Timestamp.String())
			}

			fmt.Println(text)
		}

		return nil
	},
})

func init() {
	// Command-specific flags.
	printsCmd.Flags().BoolVarP(&noTimestamps, "no-timestamps", "n", false, "omit timestamps from watch output")

	common.AddFailIfNotFoundFlag(printsCmd)
}
