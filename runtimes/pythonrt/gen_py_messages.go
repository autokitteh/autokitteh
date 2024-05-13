//go:build ignore

// Generate Python code for processing JSON from Go messages.

package main

import (
	_ "embed"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"go.autokitteh.dev/autokitteh/runtimes/pythonrt"
)

//go:embed _message.py
var header string

// Convert CamelCase to snake_case
func camelToSnake(s string) string {
	var res []rune
	for i, c := range s {
		if i > 0 && 'A' <= c && c <= 'Z' {
			res = append(res, '_')
		}
		c := unicode.ToLower(c)
		res = append(res, c)
	}
	return string(res)
}

func goToPyType(typ reflect.Type) string {
	t := typ.String()
	switch t {
	case "string":
		return "str = ''"
	case "[]uint8":
		return "bytes = b''"
	}

	switch {
	case strings.HasPrefix(t, "[]"):
		return "list = field(default_factory=list)"
	case strings.HasPrefix(t, "map["):
		return "dict = field(default_factory=dict)"
	}

	return t
}

var clsHeader = `@dataclass
class %s(Message):
`

func genPy(m pythonrt.Typed) error {
	t := reflect.TypeOf(m)
	fmt.Printf(clsHeader, t.Name())
	for _, f := range reflect.VisibleFields(t) {
		fmt.Printf("    %s: %s\n", camelToSnake(f.Name), goToPyType(f.Type))
	}
	fmt.Println()
	fmt.Printf("    def type(self): return '%s'\n", m.Type())
	fmt.Println()

	return nil
}

func main() {
	messages := []pythonrt.Typed{
		pythonrt.CallbackMessage{},
		pythonrt.ModuleMessage{},
		pythonrt.ResponseMessage{},
		pythonrt.RunMessage{},
	}

	fmt.Println(header)
	for _, m := range messages {
		genPy(m)
	}

	fmt.Println("dispatch = {")
	for _, m := range messages {
		typ := m.Type()
		name := reflect.TypeOf(m).Name()
		fmt.Printf("    '%s': %s,\n", typ, name)
	}
	fmt.Println("}")
}
