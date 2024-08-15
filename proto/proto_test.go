package proto

import (
	"embed"
	"fmt"
<<<<<<< HEAD:proto/proto_test.go
	"path/filepath"
	"regexp"
=======
	"strings"
>>>>>>> c7fa511c (fix proto validation):proto/validate_test.go
	"testing"
)

//go:embed autokitteh
var protos embed.FS

// TODO: This function is specific for v1s. If and when we have v2s in the future, this needs to be fixed.
func scan(t *testing.T, check func(t *testing.T, path, fn, dn string)) {
	dirs, err := protos.ReadDir("autokitteh")
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, dir := range dirs {
		dn := dir.Name()

		if !dir.IsDir() {
			t.Errorf("%s is not a directory, unrecognized structure", dn)
			continue
		}

		path := fmt.Sprintf("autokitteh/%s/v1", dn)

		fs, err := protos.ReadDir(path)
		if err != nil {
			t.Fatalf("%v", err)
		}

		for _, f := range fs {
			if f.IsDir() {
				t.Errorf("%s is a directory, unrecognized structure", dn)
				continue
			}

			// TODO: Once we use buf to generate, remove this
			if strings.HasSuffix(f.Name(), "remote.proto") {
				continue
			}

			t.Run(fmt.Sprintf("%s/%s", dn, f.Name()), func(t *testing.T) {
				check(t, filepath.Join(path, f.Name()), dn, f.Name())
			})
		}
	}
}

func TestParse(t *testing.T) { _ = parse(fds) }

// Make sure all protos are registered in validate.go to warm up the validator.
func TestAllValidated(t *testing.T) {
	scan(t, func(t *testing.T, _, dn, fn string) {
		found := false
		for _, fd := range fds {
			if found = fd.Path() == fmt.Sprintf("autokitteh/%s/v1/%s", dn, fn); found {
				break
			}
		}

		if !found {
			t.Errorf("%s:%s not validated, please add it to validate.go", dn, fn)
		}
	})
}

// Make sure all protos are registered in svcnames.go to warm up the validator.
func TestAllNames(t *testing.T) {
	re := regexp.MustCompile(`(?m)^service *([^ ]+) *{$`)

	scan(t, func(t *testing.T, path, fn, dn string) {
		bs, err := protos.ReadFile(path)
		if err != nil {
			t.Error("failed to read file", err)
			return
		}

		matches := re.FindAllStringSubmatch(string(bs), -1)
	L:
		for _, match := range matches {
			got := fmt.Sprintf("autokitteh.%s.v1.%s", fn, match[1])

			for _, name := range ServiceNames {
				if name == got {
					continue L
				}
			}

			t.Errorf("%s not registered in svcnames.go", got)
		}
	})
}
