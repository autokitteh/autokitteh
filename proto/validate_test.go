package proto

import (
	"embed"
	"fmt"
	"strings"
	"testing"
)

//go:embed autokitteh
var protos embed.FS

func TestParse(t *testing.T) { _ = parse(fds) }

// Make sure all protos are registered in validate.go to warm up the validator.
// TODO: This function is specific for v1s. If and when we have v2s in the future, this needs to be fixed.
func TestAllValidated(t *testing.T) {
	check := func(dn, fn string) {
		found := false
		for _, fd := range fds {
			if found = fd.Path() == fmt.Sprintf("autokitteh/%s/v1/%s", dn, fn); found {
				t.Logf("%s is validated", fmt.Sprintf("%s: %s", dn, fn))
				break
			}
		}

		if !found {
			t.Errorf("%s:%s not validated, please add it to validate.go", dn, fn)
		}
	}

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

		fs, err := protos.ReadDir(fmt.Sprintf("autokitteh/%s/v1", dn))
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

			check(dn, f.Name())
		}
	}
}
