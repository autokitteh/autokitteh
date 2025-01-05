package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type fakeDB struct {
	db.DB

	orgs       map[sdktypes.OrgID]sdktypes.Org
	users      map[sdktypes.UserID]sdktypes.User
	orgMembers map[sdktypes.OrgID]map[sdktypes.UserID]sdktypes.OrgMemberStatus
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

func (d *fakeDB) AddOrgMember(ctx context.Context, orgID sdktypes.OrgID, userID sdktypes.UserID, status sdktypes.OrgMemberStatus) error {
	if d.orgMembers == nil {
		d.orgMembers = make(map[sdktypes.OrgID]map[sdktypes.UserID]sdktypes.OrgMemberStatus)
	}

	if _, ok := d.orgMembers[orgID]; !ok {
		d.orgMembers[orgID] = make(map[sdktypes.UserID]sdktypes.OrgMemberStatus)
	}

	d.orgMembers[orgID][userID] = status
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

			assert.Equal(t, sdktypes.OrgMemberStatusActive, db.orgMembers[oid][uid])
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

			assert.Equal(t, sdktypes.OrgMemberStatusActive, db.orgMembers[oid][uid])
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

			assert.Equal(t, sdktypes.OrgMemberStatusActive, db.orgMembers[oid][uid])
		}
	}

	assert.Len(t, db.orgs, 1)
}
