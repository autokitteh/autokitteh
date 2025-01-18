package authz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/policy/opapolicy"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Basic sanity tests, more are in system tests.
// Will add more in the future as necessary.
func TestDefaultPolicy(t *testing.T) {
	tests := []struct {
		name   string
		authn  sdktypes.User // authenticated user
		id     sdktypes.ID   // resource id
		action string
		opts   []CheckOpt
		err    error
	}{
		{
			name: "no authn",
			err:  sdkerrors.ErrUnauthenticated,
		},
		{
			name:   "allow create org",
			authn:  zumi,
			id:     sdktypes.InvalidOrgID,
			action: "create:create",
			opts:   []CheckOpt{WithData("org", cats)},
		},
		{
			name:   "allow create project",
			authn:  zumi,
			id:     sdktypes.InvalidProjectID,
			action: "create:create",
			opts:   []CheckOpt{WithData("project", p)},
		},
		{
			name:   "deny create project",
			authn:  shoogy,
			id:     sdktypes.InvalidProjectID,
			action: "create:create",
			opts:   []CheckOpt{WithData("project", p)},
			err:    sdkerrors.ErrUnauthorized, // shoogy is not a member of cats.
		},
		{
			name:   "allow get user",
			authn:  shoogy,
			id:     zumi.ID(),
			action: "read:get",
		},
		{
			name:   "allow get user",
			authn:  zumi,
			id:     sufi.ID(),
			action: "read:get",
		},
		{
			name:   "allow self get user",
			authn:  zumi,
			id:     zumi.ID(),
			action: "read:get",
		},
	}

	decide, err := opapolicy.New(nil, zaptest.NewLogger(t))
	require.NoError(t, err)

	p := NewPolicyCheckFunc(zaptest.NewLogger(t), setupDB(t), decide)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := authcontext.SetAuthnUser(context.Background(), test.authn)
			err := p(ctx, test.id, test.action, test.opts...)
			assert.Equal(t, test.err, err)
		})
	}
}
