package builds_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/backend/builds"
	"go.autokitteh.dev/autokitteh/internal/backend/policy/opapolicy/opapolicytest"
	"go.autokitteh.dev/autokitteh/internal/backend/projects"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestAuthzSingleTenant(t *testing.T) {
	db, ctx := opapolicytest.InitAuthzTest(t, "!single_tenant")

	bsvc := builds.New(builds.Builds{Z: zaptest.NewLogger(t), DB: db}, nil)
	psvc := projects.New(projects.Projects{Z: zaptest.NewLogger(t), DB: db}, nil)

	// ALLOW: Simple save with no associated project.

	ctx = authcontext.SetAuthnUser(ctx, authusers.DefaultUser)

	id, err := bsvc.Save(ctx, sdktypes.NewBuild(), sdkbuildfile.EmptyBuildFileData)
	require.NoError(t, err)

	oid, err := db.GetOwner(ctx, id)
	if assert.NoError(t, err) {
		assert.Equal(t, authusers.DefaultUser.ID(), oid.ToUserID())
	}

	// ALLOW: Save with associated project which belongs to default user.

	pid, err := psvc.Create(ctx, sdktypes.NewProject())
	b := sdktypes.NewBuild().WithID(id).WithProjectID(pid)

	id, err = bsvc.Save(ctx, b, sdkbuildfile.EmptyBuildFileData)
	if assert.NoError(t, err) {
		oid, err = db.GetOwner(ctx, id)
		if assert.NoError(t, err) {
			assert.Equal(t, authusers.DefaultUser.ID(), oid.ToUserID())
		}
	}

	// ALLOW: Save with associated project which belongs to another user user.

	pid, err = psvc.Create(authcontext.SetAuthnUser(ctx, authusers.TestUser), sdktypes.NewProject())
	b = sdktypes.NewBuild().WithID(id).WithProjectID(pid)

	id, err = bsvc.Save(ctx, b, sdkbuildfile.EmptyBuildFileData)
	if assert.NoError(t, err) {
		oid, err = db.GetOwner(ctx, id)
		if assert.NoError(t, err) {
			assert.Equal(t, authusers.DefaultUser.ID(), oid.ToUserID())
		}
	}

	// ALLOW: Get with same user.

	_, err = bsvc.Get(ctx, id)
	assert.NoError(t, err)

	// ALLOW: Get with test user.

	_, err = bsvc.Get(authcontext.SetAuthnUser(ctx, authusers.TestUser), id)
	assert.NoError(t, err)
}

func TestAuthzMultiTenant(t *testing.T) {
	db, ctx := opapolicytest.InitAuthzTest(t, "!multi_tenant")

	bsvc := builds.New(builds.Builds{Z: zaptest.NewLogger(t), DB: db}, nil)
	psvc := projects.New(projects.Projects{Z: zaptest.NewLogger(t), DB: db}, nil)

	// ALLOW: Simple save with no associated project.

	ctx = authcontext.SetAuthnUser(ctx, authusers.DefaultUser)

	id, err := bsvc.Save(ctx, sdktypes.NewBuild(), sdkbuildfile.EmptyBuildFileData)
	require.NoError(t, err)

	oid, err := db.GetOwner(ctx, id)
	if assert.NoError(t, err) {
		assert.Equal(t, authusers.DefaultUser.ID(), oid.ToUserID())
	}

	// ALLOW: Save with associated project which belongs to default user.

	pid, err := psvc.Create(ctx, sdktypes.NewProject())
	b := sdktypes.NewBuild().WithID(id).WithProjectID(pid)

	id, err = bsvc.Save(ctx, b, sdkbuildfile.EmptyBuildFileData)
	if assert.NoError(t, err) {
		oid, err = db.GetOwner(ctx, id)
		if assert.NoError(t, err) {
			assert.Equal(t, authusers.DefaultUser.ID(), oid.ToUserID())
		}
	}

	// DENY: Save with associated project which belongs to another user user.

	pid, err = psvc.Create(authcontext.SetAuthnUser(ctx, authusers.TestUser), sdktypes.NewProject())
	if assert.NoError(t, err) {
		b = sdktypes.NewBuild().WithID(id).WithProjectID(pid)
		_, err = bsvc.Save(ctx, b, sdkbuildfile.EmptyBuildFileData)
		assert.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	}

	// ALLOW: Get with same user.

	_, err = bsvc.Get(ctx, id)
	assert.NoError(t, err)

	// DENY: Get with test user.

	_, err = bsvc.Get(authcontext.SetAuthnUser(ctx, authusers.TestUser), id)
	assert.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
}
