package common

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func OpenURL(cmd *cobra.Command, link string) error {
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		var err error

		if link, err = url.JoinPath(ServerURL().String(), link); err != nil {
			return err
		}
	}

	fmt.Fprintf(
		cmd.ErrOrStderr(),
		`Attempting to automatically open a link using your default browser.
If the browser does not open, please open the following URL:

%s`+"\n\n", link)

	_ = browser.OpenURL(link)

	return nil
}
