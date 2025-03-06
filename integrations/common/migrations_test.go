package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestRenameVar(t *testing.T) {
	vars := newFakeVars()
	vsid := sdktypes.NewVarScopeID(sdktypes.NewConnectionID())
	ctx := context.Background()

	v := sdktypes.NewVar(sdktypes.NewSymbol("a")).SetValue("a")
	require.NoError(t, vars.Set(ctx, v.WithScopeID(vsid)))
	v = sdktypes.NewVar(sdktypes.NewSymbol("b")).SetValue("b")
	require.NoError(t, vars.Set(ctx, v.WithScopeID(vsid)))

	err := RenameVar(ctx, vars, vsid, sdktypes.NewSymbol("b"), sdktypes.NewSymbol("c"))
	require.NoError(t, err)

	v = vars.data[vsid][sdktypes.NewSymbol("a")]
	assert.True(t, v.IsValid())
	v = vars.data[vsid][sdktypes.NewSymbol("b")]
	assert.False(t, v.IsValid())
	v = vars.data[vsid][sdktypes.NewSymbol("c")]
	assert.True(t, v.IsValid())
}

func TestMigrateAuthType(t *testing.T) {
	vars := newFakeVars()
	tests := []struct {
		initial string
		want    string
	}{
		{
			initial: integrations.OAuth,
			want:    integrations.OAuthDefault,
		},
		{
			initial: integrations.OAuthDefault,
			want:    integrations.OAuthDefault,
		},
		{
			initial: "other",
			want:    "other",
		},
	}
	for _, tt := range tests {
		t.Run(tt.initial, func(t *testing.T) {
			vsid := sdktypes.NewVarScopeID(sdktypes.NewConnectionID())
			ctx := context.Background()

			v := sdktypes.NewVar(AuthTypeVar).SetValue(tt.initial)
			require.NoError(t, vars.Set(ctx, v.WithScopeID(vsid)))

			require.NoError(t, MigrateAuthType(ctx, vars, vsid))

			assert.Equal(t, tt.want, vars.data[vsid][AuthTypeVar].Value())
		})
	}
}

func TestMigrateDateTimeToRFC3339(t *testing.T) {
	vars := newFakeVars()
	tests := []struct {
		name    string
		initial string
		want    string
		wantErr bool
	}{
		{
			name:    "already_rfc3339",
			initial: "2025-02-28T11:22:33Z",
			want:    "2025-02-28T11:22:33Z",
		},
		{
			name:    "time_string",
			initial: "2025-02-28 11:22:33 -0800",
			want:    "2025-02-28T19:22:33Z",
		},
		{
			name:    "invalid_format",
			initial: "invalid_format",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vsid := sdktypes.NewVarScopeID(sdktypes.NewConnectionID())
			ctx := context.Background()

			v := sdktypes.NewVar(OAuthExpiryVar).SetValue(tt.initial)
			require.NoError(t, vars.Set(ctx, v.WithScopeID(vsid)))

			err := MigrateDateTimeToRFC3339(ctx, vars, vsid, OAuthExpiryVar)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, vars.data[vsid][OAuthExpiryVar].Value())
		})
	}
}
