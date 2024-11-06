package sdktypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventDestinationID(t *testing.T) {
	tests := []struct {
		typeID         string
		wantConnection bool
		wantTrigger    bool
	}{
		{
			typeID:         "con_01jbw762gzfe4vbv30zvf6f4cj",
			wantConnection: true,
		},
		{
			typeID:      "trg_01jbw762h5f8mvte837j1qqtfr",
			wantTrigger: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.typeID, func(t *testing.T) {
			id, err := ParseEventDestinationID(tt.typeID)
			assert.NoError(t, err)

			if tt.wantConnection {
				assert.True(t, id.IsConnectionID())
			} else {
				assert.False(t, id.IsConnectionID())
			}

			if tt.wantTrigger {
				assert.True(t, id.IsTriggerID())
			} else {
				assert.False(t, id.IsTriggerID())
			}

			assert.Equal(t, tt.typeID, id.String())
		})
	}
}
