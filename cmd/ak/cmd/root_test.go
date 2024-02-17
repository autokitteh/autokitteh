package cmd

import (
	"reflect"
	"testing"
)

func Test_parseConfigs(t *testing.T) {
	tests := []struct {
		name    string
		pairs   []string
		want    map[string]any
		wantErr bool
	}{
		{
			name:  "success",
			pairs: []string{"foo=bar"},
			want:  map[string]any{"foo": "bar"},
		},
		{
			name:    "failure",
			pairs:   []string{"foobar"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConfigs(tt.pairs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseConfigs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
