package authz

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/policy"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	policyRootPath                        = "/authz/"
	useAuthnForDefaultListFilterOwnerPath = policyRootPath + "use_authn_for_default_list_filter_owner"
)

// GlobalSettings contains global settings that are fetched from the policy
// and not from the server configuration.
type GlobalSettings struct {
	// When filter owner is not set, should use the authenticated user as the filter owner?
	UseAuthnForDefaultListFilterOwner bool
}

func GetGlobals(ctx context.Context, d policy.DecideFunc) (*GlobalSettings, error) {
	v, err := d(ctx, useAuthnForDefaultListFilterOwnerPath, nil)
	if err != nil {
		return nil, err
	}

	u, ok := v.(bool)
	if !ok {
		return nil, fmt.Errorf(useAuthnForDefaultListFilterOwnerPath + ": not a boolean")
	}

	return &GlobalSettings{
		UseAuthnForDefaultListFilterOwner: u,
	}, nil
}

func (g GlobalSettings) DefaultOwnerFromAuthn(ctx context.Context) sdktypes.OwnerID {
	if !g.UseAuthnForDefaultListFilterOwner {
		return sdktypes.InvalidOwnerID
	}

	return sdktypes.NewOwnerID(authcontext.GetAuthnInferredUserID(ctx))
}
