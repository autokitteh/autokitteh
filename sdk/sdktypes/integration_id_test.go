package sdktypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamedID(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"test", "int_3kth00testaf9d33c5697341f0"},
		{"testing", "int_3kthtest1n7bf767860ab68268"},
		{"123abc", "int_3kth123abc819fc78d4caa616c"},
		{"iou", "int_3kth00010v08d1729ff0ad290d"},
		{"foo-bar", "int_3kthf000ba732f13282aa60744"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewIntegrationIDFromName(tt.name)
			assert.Equal(t, tt.want, id.String())
		})
	}
}
