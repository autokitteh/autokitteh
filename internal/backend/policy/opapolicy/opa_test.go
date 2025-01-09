package opapolicy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	policyContent = `
-- default/policy.rego --
package policy

import rego.v1

passthrough := input.passthrough
`

	testfs = kittehs.Must1(kittehs.TxtarStringToFS(policyContent))
)

func TestPassthrough(t *testing.T) {
	d, err := New(&Config{fs: testfs}, zaptest.NewLogger(t))
	require.NoError(t, err)
	require.NotNil(t, d)

	u := sdktypes.NewUser().WithNewID().WithDisplayName("meow").WithStatus(sdktypes.UserStatusActive)

	tests := []struct {
		name     string
		in       any
		expected any
	}{
		{
			name:     "passthrough",
			in:       map[string]any{"passthrough": "passthrough"},
			expected: "passthrough",
		},
		{
			name: "passthrough",
			in: map[string]any{"passthrough": map[string]any{
				"bool":  true,
				"int":   42,
				"slice": []any{true, 42, "meow"},
				"user":  u,
			}},
			expected: map[string]any{
				"bool":  true,
				"int":   json.Number("42"),
				"slice": []any{true, json.Number("42"), "meow"},
				"user": map[string]any{
					"user_id":      u.ID().String(),
					"display_name": u.DisplayName(),
					"status":       "USER_STATUS_ACTIVE",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := d(context.Background(), "policy/"+test.name, test.in)
			if assert.NoError(t, err) {
				assert.Equal(t, test.expected, v)
			}
		})
	}
}
