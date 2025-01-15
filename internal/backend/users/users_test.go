package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/orgs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type fakeDB struct {
	db.DB

	orgs       map[sdktypes.OrgID]sdktypes.Org
	users      map[sdktypes.UserID]sdktypes.User
	orgMembers map[sdktypes.OrgID]map[sdktypes.UserID]sdktypes.OrgMember
}

func (d *fakeDB) Transaction(ctx context.Context, f func(tx db.DB) error) error { return f(d) }

func (d *fakeDB) CreateOrg(ctx context.Context, org sdktypes.Org) (sdktypes.OrgID, error) {
	if d.orgs == nil {
		d.orgs = make(map[sdktypes.OrgID]sdktypes.Org)
	}

	id := sdktypes.NewOrgID()
	d.orgs[id] = org.WithID(id)
	return id, nil
}

func (d *fakeDB) CreateUser(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error) {
	if d.users == nil {
		d.users = make(map[sdktypes.UserID]sdktypes.User)
	}

	id := sdktypes.NewUserID()
	d.users[id] = user.WithID(id)
	return id, nil
}

func (d *fakeDB) AddOrgMember(ctx context.Context, m sdktypes.OrgMember) error {
	oid, uid := m.OrgID(), m.UserID()

	if d.orgMembers == nil {
		d.orgMembers = make(map[sdktypes.OrgID]map[sdktypes.UserID]sdktypes.OrgMember)
	}

	if _, ok := d.orgMembers[oid]; !ok {
		d.orgMembers[oid] = make(map[sdktypes.UserID]sdktypes.OrgMember)
	}

	d.orgMembers[oid][uid] = m
	return nil
}

func TestCreateWithoutDefaultOrgInCfg(t *testing.T) {
	db := &fakeDB{}
	us, err := New(&Config{}, db, zaptest.NewLogger(t))
	require.NoError(t, err)

	ctx := authcontext.SetAuthnSystemUser(context.Background())

	uid, err := us.Create(ctx, sdktypes.NewUser().WithEmail("someone@somewhere"))
	if assert.NoError(t, err) {
		assert.True(t, uid.IsValid())

		u := db.users[uid]
		if assert.True(t, u.IsValid()) {
			assert.Equal(t, "someone@somewhere", u.Email())

			oid := u.DefaultOrgID()
			if assert.True(t, oid.IsValid()) {
				assert.True(t, db.orgs[oid].IsValid())
			}

			assert.Equal(t, sdktypes.OrgMemberStatusActive, db.orgMembers[oid][uid].Status())
			assert.Equal(t, []sdktypes.Symbol{orgs.OrgAdminRoleName}, db.orgMembers[oid][uid].Roles())
		}
	}

	assert.Len(t, db.orgs, 1)
}

func TestCreateWithDefaultOrgInCfg(t *testing.T) {
	db := &fakeDB{}
	ctx := authcontext.SetAuthnSystemUser(context.Background())

	oid, err := db.CreateOrg(ctx, sdktypes.NewOrg())
	require.NoError(t, err)

	us, err := New(&Config{DefaultOrgID: oid.String()}, db, zaptest.NewLogger(t))
	require.NoError(t, err)

	uid, err := us.Create(ctx, sdktypes.NewUser().WithEmail("someone@somewhere"))
	if assert.NoError(t, err) {
		assert.True(t, uid.IsValid())

		u := db.users[uid]
		if assert.True(t, u.IsValid()) {
			assert.Equal(t, "someone@somewhere", u.Email())
			assert.Equal(t, oid, u.DefaultOrgID())
			assert.Equal(t, sdktypes.OrgMemberStatusActive, db.orgMembers[oid][uid].Status())
			assert.Len(t, db.orgMembers[oid][uid].Roles(), 0)
			assert.Equal(t, sdktypes.UserStatusActive, u.Status())
		}
	}

	assert.Len(t, db.orgs, 1)
}

func TestCreateWithDefaultOrgInUser(t *testing.T) {
	db := &fakeDB{}
	ctx := authcontext.SetAuthnSystemUser(context.Background())

	oid, err := db.CreateOrg(ctx, sdktypes.NewOrg())
	require.NoError(t, err)

	us, err := New(&Config{}, db, zaptest.NewLogger(t))
	require.NoError(t, err)

	uid, err := us.Create(ctx, sdktypes.NewUser().WithEmail("someone@somewhere").WithDefaultOrgID(oid))
	if assert.NoError(t, err) {
		assert.True(t, uid.IsValid())

		u := db.users[uid]
		if assert.True(t, u.IsValid()) {
			assert.Equal(t, "someone@somewhere", u.Email())
			assert.Equal(t, oid, u.DefaultOrgID())
			assert.Equal(t, sdktypes.OrgMemberStatusActive, db.orgMembers[oid][uid].Status())
			assert.Len(t, db.orgMembers[oid][uid].Roles(), 0)
			assert.Equal(t, sdktypes.UserStatusActive, u.Status())
		}
	}

	assert.Len(t, db.orgs, 1)
}

func TestCreateInvitedUser(t *testing.T) {
	db := &fakeDB{}
	ctx := authcontext.SetAuthnSystemUser(context.Background())

	us, err := New(&Config{}, db, zaptest.NewLogger(t))
	require.NoError(t, err)

	uid, err := us.Create(ctx, sdktypes.NewUser().WithEmail("someone@somewhere").WithStatus(sdktypes.UserStatusInvited))
	if assert.NoError(t, err) {
		assert.True(t, uid.IsValid())

		u := db.users[uid]
		if assert.True(t, u.IsValid()) {
			assert.Equal(t, "someone@somewhere", u.Email())
			assert.True(t, u.DefaultOrgID().IsValid())
			assert.Equal(t, sdktypes.OrgMemberStatusActive, db.orgMembers[u.DefaultOrgID()][uid].Status())
			assert.Equal(t, sdktypes.UserStatusInvited, u.Status())
		}
	}

	assert.Len(t, db.orgs, 1)
}
