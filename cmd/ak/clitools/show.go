package clitools

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
)

func show(what interface{}) string {
	marshal := json.Marshal

	if Settings.Yaml {
		marshal = yaml.Marshal
	} else if Settings.IndentJSON {
		marshal = func(what interface{}) ([]byte, error) { return json.MarshalIndent(what, "", "  ") }
	}

	bs, err := marshal(what)
	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(bs))
}

func ShowStderr(what interface{}) { fmt.Fprintln(os.Stderr, show(what)) }
func Show(what interface{})       { fmt.Println(show(what)) }
