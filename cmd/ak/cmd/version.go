package cmd

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/version"
)

type Version struct {
	Version   string           `json:"version"`
	Commit    string           `json:"commit,omitempty"`
	Date      string           `json:"date,omitempty"`
	User      string           `json:"user,omitempty"`
	Host      string           `json:"host,omitempty"`
	BuildInfo *debug.BuildInfo `json:"buildinfo,omitempty"`
}

func (v Version) String() string {
	var w strings.Builder
	w.WriteString(v.Version)

	if v.Commit != "" {
		w.WriteString(" " + v.Commit)
	}
	if v.Date != "" {
		w.WriteString(" " + v.Date)
	}
	if v.User != "" {
		w.WriteString(" by " + v.User)
	}
	if v.Host != "" {
		w.WriteString(" on " + v.Host)
	}
	if v.BuildInfo != nil {
		w.WriteString(fmt.Sprintf("\n\n%v", v.BuildInfo))
	}
	return w.String()
}

var full, build bool

var versionCmd = common.StandardCommand(&cobra.Command{
	Use:     "version [--full] [--build]",
	Short:   "Print CLI version information",
	Aliases: []string{"ver", "v"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		data := Version{Version: version.Version}

		if full {
			data.Commit = version.Commit
			data.Date = version.Time
			data.User = version.User
			data.Host = version.Host
		}
		if build {
			info, ok := debug.ReadBuildInfo()
			if !ok {
				return errors.New("unable to read build info")
			}
			data.BuildInfo = info
		}

		common.Render(data)
		return nil
	},
})

func init() {
	// Command-specific flags.
	versionCmd.Flags().BoolVarP(&full, "full", "f", false, "print code commit details")
	versionCmd.Flags().BoolVarP(&build, "build", "b", false, "print verbose Go build details")
}
