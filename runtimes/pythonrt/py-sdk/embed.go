package pysdk

import (
	"embed"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed autokitteh/*.py
var codeFS embed.FS

//go:embed pyproject.toml
var pyProject string

var (
	clientDefRegex = regexp.MustCompile(`^def (\w+_client)\(`)

	clientNames = func() (names []string) {
		kittehs.Must0(fs.WalkDir(codeFS, "autokitteh", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || !strings.HasSuffix(path, ".py") {
				return nil
			}

			bs, err := fs.ReadFile(codeFS, path)
			if err != nil {
				return err
			}

			for line := range strings.SplitSeq(string(bs), "\n") {
				line = strings.TrimSpace(line)
				matches := clientDefRegex.FindStringSubmatch(line)
				if len(matches) > 1 {
					names = append(names, matches[1])
				}
			}

			return nil
		}))

		sort.Strings(names)

		return
	}()
)

func ClientNames() []string { return clientNames }

func Dependencies() (names []string) {
	var data struct {
		Project struct {
			OptionalDependencies struct {
				All []string `toml:"all"`
			} `toml:"optional-dependencies"`
		} `toml:"project"`
	}

	_ = kittehs.Must1(toml.Decode(pyProject, &data))

	names = kittehs.Transform(data.Project.OptionalDependencies.All, func(dep string) string {
		return strings.SplitN(dep, " ", 2)[0]
	})

	sort.Strings(names)

	return
}
