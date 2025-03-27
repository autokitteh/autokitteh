package telemetry

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestParseAttributes(t *testing.T) {
	tests := []struct {
		name      string
		attrs     []string
		expected  []attribute.KeyValue
		shouldErr bool
	}{
		{
			name:     "empty attributes",
			attrs:    []string{},
			expected: nil,
		},
		{
			name:     "single pair",
			attrs:    []string{"key1", "value1"},
			expected: []attribute.KeyValue{attribute.String("key1", "value1")},
		},
		{
			name:     "multiple pairs",
			attrs:    []string{"key1", "value1", "key2", "value2"},
			expected: []attribute.KeyValue{attribute.String("key1", "value1"), attribute.String("key2", "value2")},
		},
		{
			name:      "odd number of items",
			attrs:     []string{"key1", "value1", "key2"},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrs, err := parseAttributes(tt.attrs)
			if tt.shouldErr && err != nil {
				return // Expected error
			}

			if len(attrs) != len(tt.expected) {
				t.Errorf("Expected %d attributes, got %d", len(tt.expected), len(attrs))
				return
			}

			require.Equal(t, tt.expected, attrs)
		})
	}
}
