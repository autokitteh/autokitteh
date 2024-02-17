package webhooks

import (
	"testing"

	"go.uber.org/zap"
)

type installation struct {
	ID *int64
}

type badEvent struct {
	NoInstallationField *int64
}

type goodEvent struct {
	Installation *installation
}

func TestExtractInstallationID(t *testing.T) {
	id := int64(1)
	tests := []struct {
		name  string
		event any
		want  string
	}{
		{
			name:  "bad_event",
			event: badEvent{},
			want:  "",
		},
		{
			name: "good_event",
			event: goodEvent{
				Installation: &installation{ID: &id},
			},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractInstallationID(zap.L(), tt.event, "eventType"); got != tt.want {
				t.Errorf("extractInstallationID() = %v, want %v", got, tt.want)
			}
		})
	}
}
