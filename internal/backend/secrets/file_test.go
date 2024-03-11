package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestFileSecretsSet(t *testing.T) {
	sec, _ := NewFakeSecrets(zap.L())

	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "first_set_works_as_expected",
			data: map[string]string{"key1": "value1"},
		},
		{
			name: "second_set_overwrites_first",
			data: map[string]string{"key2": "value2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := sec.Set("prefix", "name", tt.data); (err != nil) != tt.wantErr {
				t.Errorf("fileSecrets.Set() error = %v, wantErr %v", err, false)
			}

			got, err := sec.Get("prefix", "name")
			if err != nil {
				t.Fatalf("fileSecrets.Get() error = %v, wantErr %v", err, false)
			}
			assert.Equal(t, tt.data, got, "fileSecrets.Get() = %v, want %v", got, tt.data)
		})
	}
}

func TestFileSecretsGet(t *testing.T) {
	tests := []struct {
		name    string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "get_nonexistent_returns_nil_not_error",
			want: nil,
		},
		{
			name: "get_works_as_expected",
			want: map[string]string{"key": "value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sec, _ := NewFakeSecrets(zap.L())

			if tt.want != nil {
				if err := sec.Set("prefix", tt.name, tt.want); err != nil {
					t.Fatalf("fileSecrets.Set() error = %v, wantErr %v", err, false)
				}
			}

			got, err := sec.Get("prefix", tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("fileSecrets.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got, "fileSecrets.Get() = %v, want %v", got, tt.want)
		})
	}
}

func TestFileSecretsDelete(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		wantErr bool
	}{
		{
			name: "destroy_nonexistent_secret_isnt_an_error",
			data: nil,
		},
		{
			name: "get_works_as_expected",
			data: map[string]string{"key": "value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sec, _ := NewFakeSecrets(zap.L())

			if tt.data != nil {
				if err := sec.Set("prefix", tt.name, tt.data); err != nil {
					t.Fatalf("fileSecrets.Set() error = %v, wantErr %v", err, false)
				}
			}

			err := sec.Delete("prefix", tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("fileSecrets.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := sec.Get("prefix", tt.name)
			if err != nil {
				t.Fatalf("fileSecrets.Get() error = %v, wantErr %v", err, false)
			}
			if got != nil {
				t.Errorf("fileSecrets.Get() = %v, want nil", got)
			}
		})
	}
}
