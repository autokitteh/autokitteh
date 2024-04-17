package sdktypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamedID(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"test", "int_3kth00test8c093f7e9fccbf69"},
		{"testing", "int_3kthtest1nffd1afe03e3d3dff"},
		{"123abc", "int_3kth123abc0fd730e26b553e85"},
		{"iou", "int_3kth00010vd8c1a7186b98aa02"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewIntegrationIDFromName(tt.name)
			assert.Equal(t, tt.want, id.String())
		})
	}
}
