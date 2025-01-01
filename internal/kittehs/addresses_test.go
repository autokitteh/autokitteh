package kittehs

import (
	"testing"
)

func TestBindingAddress(t *testing.T) {
	input := "1234"
	got := BindingAddress("1234")
	want := "0.0.0.0:1234"
	if got != want {
		t.Errorf("BindingAddress(%q) = %q, want %q", input, got, want)
	}
}

func TestDisplayAddress(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"0.0.0.0:1", "localhost:1"},
		{"[::]:22", "localhost:22"},
		{"[::1]:333", "[::1]:333"},
		{"127.0.0.1:333", "127.0.0.1:333"},
		{"localhost:4444", "localhost:4444"},
		{"host.com:55555", "host.com:55555"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := DisplayAddress(test.input); got != test.want {
				t.Errorf("DisplayAddress(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
