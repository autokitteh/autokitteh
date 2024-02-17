package common

import (
	"encoding/json"
	"fmt"
	"io"
)

type Renderer func(any)

type Texter interface{ Text() string }

var (
	// If we ever decide to make rendering stateful, i.e. pass a reference of the
	// calling command in [Render] calls, we could use the command's [OutOrStdout]
	// and [ErrOrStderr] functions instead of these global variables.
	outOrStdout, errOrStderr io.Writer

	renderer Renderer = TextRenderer
)

func GetWriters() (io.Writer, io.Writer) {
	return outOrStdout, errOrStderr
}

func SetWriters(out, err io.Writer) {
	outOrStdout = out
	errOrStderr = err
}

func SetRenderer(r Renderer) { renderer = r }

func TextRenderer(o any) {
	var out string

	if oo, ok := o.(Texter); ok {
		out = oo.Text()
	} else if _, ok := o.(fmt.Stringer); ok {
		out = fmt.Sprintf("%v", o)
	} else {
		NiceJSONRenderer(o)
		return
	}

	if out != "" {
		fmt.Fprintln(outOrStdout, out)
	}
}

func JSONRenderer(o any) {
	text, err := json.Marshal(o)
	if err != nil {
		renderError(fmt.Errorf("marshal: %w", err))
		return
	}

	fmt.Fprintln(outOrStdout, string(text))
}

func NiceJSONRenderer(o any) {
	text, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		renderError(fmt.Errorf("marshal: %w", err))
		return
	}

	fmt.Fprintln(outOrStdout, string(text))
}

func renderError(err error) {
	fmt.Fprintf(errOrStderr, "error: %v", err)
}

func Render(o any) { renderer(o) }

type KV struct {
	K string
	V any
}

func (kv KV) Text() string                 { return fmt.Sprintf("%s: %v", kv.K, kv.V) }
func (kv KV) MarshalJSON() ([]byte, error) { return json.Marshal(map[string]any{kv.K: kv.V}) }

var (
	_ Texter         = KV{}
	_ json.Marshaler = KV{}
)

func RenderKV(k string, v any) { Render(KV{K: k, V: v}) }

// This will not print anything if V is nil for text rendering.
// Should make output parsing easier for get operations.
// TODO: should we even care for non-json output parsing?
type KVIfV[T any] struct {
	K string
	V *T
}

func (kv KVIfV[T]) Text() string {
	if kv.V == nil {
		return ""
	}

	return fmt.Sprintf("%s: %v", kv.K, kv.V)
}

func (kv KVIfV[T]) MarshalJSON() ([]byte, error) { return json.Marshal(map[string]any{kv.K: kv.V}) }

var (
	_ Texter         = KV{}
	_ json.Marshaler = KV{}
)

func RenderKVIfV[T any](k string, v *T) {
	Render(KVIfV[T]{K: k, V: v})
}

func RenderList[T fmt.Stringer](ts []T) {
	for _, t := range ts {
		Render(t)
	}
}
