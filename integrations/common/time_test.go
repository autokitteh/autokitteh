package common

import (
	"testing"
	"time"
)

func TestParseGoTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		ts      string
		want    time.Time
		wantErr bool
	}{
		{
			name: "minimal",
			ts:   "2006-01-02 15:04:05 -0700",
			want: time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("", -7*60*60)),
		},
		{
			name: "go_playground",
			ts:   "2009-11-10 23:00:00 +0000 UTC m=+0.000000001",
			want: time.Date(2009, 11, 10, 23, 0, 0, 0, time.FixedZone("", 0)),
		},
		{
			name: "with_milliseconds",
			ts:   "2009-11-10 23:00:00.123 +0000 UTC m=+0.000000001",
			want: time.Date(2009, 11, 10, 23, 0, 0, 0, time.FixedZone("", 0)),
		},
		{
			name: "with_nanoseconds",
			ts:   "2025-02-28 11:22:33.123456789 -0800 PST m=+0.000199937",
			want: time.Date(2025, 2, 28, 11, 22, 33, 123456789, time.FixedZone("", -8*60*60)),
		},
		{
			name:    "rfc_3339",
			ts:      "2023-10-10T10:10:10Z",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGoTimestamp(tt.ts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGoTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Format(time.RFC3339) != tt.want.Format(time.RFC3339) {
				t.Errorf("ParseGoTimestamp() = %q, want %q", got, tt.want)
			}
		})
	}
}
