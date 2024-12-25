package systest

import "testing"

func TestNextField(t *testing.T) {
	tests := []struct {
		text string
		a, b string
	}{
		{},
		{`"meow world" moo`, "meow world", " moo"},
		{`'meow world' moo`, "meow world", " moo"},
		{`'meow "world"' moo`, `meow "world"`, " moo"},
		{`meow world moo`, "meow", "world moo"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			a, b := nextField(tt.text)
			if a != tt.a || b != tt.b {
				t.Errorf("got %q, %q; want %q, %q", a, b, tt.a, tt.b)
			}
		})
	}
}
