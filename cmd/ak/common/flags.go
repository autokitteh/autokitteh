package common

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func AddFailIfNotFoundFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("fail", "f", false, "fail if not found")
}

func AddFailIfError(cmd *cobra.Command) {
	cmd.Flags().BoolP("fail", "f", false, "fail on error")
}

// keep given error, if passed or return notFound if !found condition
func AddNotFoundErrIfCond(err error, found bool) error {
	if err == nil && !found {
		err = sdkerrors.ErrNotFound
	}
	return err
}

// ToExitCodeWithSkipNotFoundFlag returns the given command's error (may be nil) with an OS
// exit code, but considers the "--fail" flag: if set to false, we skip "not found" errors.
func ToExitCodeWithSkipNotFoundFlag(cmd *cobra.Command, err error, whats ...string) error {
	if err == nil {
		return nil
	}

	if ToExitCode(err) == NotFoundExitCode {
		flags := cmd.Flags()
		if flags.Lookup("fail") != nil && !kittehs.Must1(flags.GetBool("fail")) {
			return nil
		}
	}

	return WrapError(err, whats...)
}
