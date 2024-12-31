package dbgorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/dbtime"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestUpdatedFields(t *testing.T) {
	dbtime.Freeze()

	ctx := context.Background()

	u := sdktypes.NewUser().WithNewID().WithDisplayName("DISPLAY_NAME")

	m, err := updatedFields(ctx, u, nil)
	if assert.NoError(t, err) {
		// returns only the mutable fields + updated_at.
		assert.Equal(t, map[string]any{
			"display_name":   "DISPLAY_NAME",
			"default_org_id": "",
			"disabled":       false,
			"updated_at":     dbtime.Now().UTC(),
		}, m)
	}

	m, err = updatedFields(ctx, u, kittehs.Must1(fieldmaskpb.New(&sdktypes.UserPB{}, "disabled", "display_name")))
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]any{
			"disabled":     false,
			"display_name": "DISPLAY_NAME",
			"updated_at":   dbtime.Now().UTC(),
		}, m)
	}

	// immutable.
	_, err = updatedFields(ctx, u, kittehs.Must1(fieldmaskpb.New(&sdktypes.UserPB{}, "user_id")))
	assert.ErrorAs(t, err, &sdkerrors.ErrInvalidArgument{})

	// mutable only by updatedFields, not by the caller.
	_, err = updatedFields(ctx, u, &fieldmaskpb.FieldMask{Paths: []string{"updated_at"}})
	assert.ErrorAs(t, err, &sdkerrors.ErrInvalidArgument{})

	ctx = authcontext.SetAuthnUser(ctx, authusers.DefaultUser)
	m, err = updatedFields(ctx, u, kittehs.Must1(fieldmaskpb.New(&sdktypes.UserPB{}, "disabled", "display_name")))
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]any{
			"display_name": "DISPLAY_NAME",
			"disabled":     false,
			"updated_at":   dbtime.Now().UTC(),
			"updated_by":   authusers.DefaultUser.ID().UUIDValue(),
		}, m)
	}
}
