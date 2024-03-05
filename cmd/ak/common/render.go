package common

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

type Renderer func(any)

type Texter interface{ Text() string }

var renderer Renderer = TextRenderer

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
		fmt.Fprintln(os.Stdout, out)
	}
}

func JSONRenderer(o any) {
	text, err := json.Marshal(o)
	if err != nil {
		renderError(fmt.Errorf("marshal: %w", err))
		return
	}

	fmt.Fprintln(os.Stdout, string(text))
}

func NiceJSONRenderer(o any) {
	text, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		renderError(fmt.Errorf("marshal: %w", err))
		return
	}

	fmt.Fprintln(os.Stdout, string(text))
}

func renderError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v", err)
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
	V T
}

func (kv KVIfV[T]) Text() string {
	if reflect.ValueOf(kv.V).IsZero() {
		return ""
	}

	return fmt.Sprintf("%s: %v", kv.K, kv.V)
}

func (kv KVIfV[T]) MarshalJSON() ([]byte, error) { return json.Marshal(map[string]any{kv.K: kv.V}) }

var (
	_ Texter         = KV{}
	_ json.Marshaler = KV{}
)

func RenderKVIfV[T any](k string, v T) {
	Render(KVIfV[T]{K: k, V: v})
}

func RenderList[T fmt.Stringer](ts []T) {
	for _, t := range ts {
		Render(t)
	}
}
