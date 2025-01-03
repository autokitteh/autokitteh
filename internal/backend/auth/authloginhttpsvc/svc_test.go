package authloginhttpsvc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktest"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var testUser = sdktypes.NewUser().WithNewID().WithDisplayName("test").WithEmail("someone@somewhere").WithStatus(sdktypes.UserStatusActive)

func isErrHandler(h http.Handler) bool { _, ok := h.(errorHandler); return ok }
func errHandlerCode(h http.Handler) int {
	if !isErrHandler(h) {
		return 0
	}

	return h.(errorHandler).code
}

func TestNewSuccessLoginHandlerImmediate(t *testing.T) {
	users := sdktest.TestUsers{IDGen: testUser.ID}

	s := &svc{
		Deps: Deps{
			Users: &users,
			Cfg: &Config{
				RejectNewUsers: true,
			},
			L: zap.NewNop(),
		},
	}

	ld := &loginData{
		ProviderName: "test",
		DisplayName:  "test",
	}

	ctx := context.TODO()

	// cannot login without email.
	h := s.newSuccessLoginHandler(ctx, ld)
	assert.Equal(t, http.StatusBadRequest, errHandlerCode(h))

	ld.Email = "someone@somewhere"

	assertCounts := func(create, update, get int) {
		assert.Equal(t, get, users.GetCalledCount)
		assert.Equal(t, create, users.CreateCalledCount)
		assert.Equal(t, update, users.UpdateCalledCount)
	}

	// reject new users.
	h = s.newSuccessLoginHandler(ctx, ld)
	assert.Equal(t, http.StatusForbidden, errHandlerCode(h))
	assertCounts(0, 0, 1)

	// new user.
	users.Reset()
	s.Deps.Cfg.RejectNewUsers = false

	h = s.newSuccessLoginHandler(ctx, ld)
	if assert.False(t, isErrHandler(h), h) {
		assert.Equal(
			t,
			testUser,
			users.Users[testUser.ID()],
		)
	}
	assertCounts(1, 0, 1)

	// user already exists and is active.
	users.Reset(testUser)

	h = s.newSuccessLoginHandler(ctx, ld)
	if assert.False(t, isErrHandler(h), h) {
		assert.Equal(
			t,
			testUser,
			users.Users[testUser.ID()],
		)
	}
	assertCounts(0, 0, 1)

	// invited user.
	users.Reset(testUser.WithStatus(sdktypes.UserStatusInvited).WithDisplayName(""))

	h = s.newSuccessLoginHandler(ctx, ld)
	if assert.False(t, isErrHandler(h), h) {
		assert.Equal(
			t,
			testUser,
			users.Users[testUser.ID()],
		)
	}
	assertCounts(0, 1, 1)

	// invited user should work even if rejecting new users.
	users.Reset(testUser.WithStatus(sdktypes.UserStatusInvited).WithDisplayName(""))

	s.Deps.Cfg.RejectNewUsers = true

	h = s.newSuccessLoginHandler(ctx, ld)
	if assert.False(t, isErrHandler(h), h) {
		assert.Equal(
			t,
			testUser,
			users.Users[testUser.ID()],
		)
	}
	assertCounts(0, 1, 1)

	// disabled user.
	users.Reset(testUser.WithStatus(sdktypes.UserStatusDisabled))
	h = s.newSuccessLoginHandler(ctx, ld)
	assert.Equal(t, http.StatusForbidden, errHandlerCode(h))
	assertCounts(0, 0, 1)
}

func TestNewSuccessLoginHandlerSessions(t *testing.T) {
	users := sdktest.TestUsers{
		Users: map[sdktypes.UserID]sdktypes.User{
			testUser.ID(): testUser,
		},
	}

	sessions := kittehs.Must1(authsessions.New(authsessions.Configs.Dev))

	s := &svc{
		Deps: Deps{
			Sessions: sessions,
			Users:    &users,
			Cfg:      &Config{},
			L:        zap.NewNop(),
		},
	}

	ld := &loginData{
		ProviderName: "test",
		DisplayName:  "test",
		Email:        "someone@somewhere",
	}

	// new user.
	h := s.newSuccessLoginHandler(context.TODO(), ld)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/?noredir=1", nil))
	if assert.Equal(t, http.StatusOK, w.Code) {
		if cookies := w.Result().Cookies(); assert.Len(t, cookies, 2) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.AddCookie(cookies[0])
			r.AddCookie(cookies[1])

			sd, err := sessions.Get(r)
			if assert.NoError(t, err) {
				assert.Equal(t, testUser.ID(), sd.UserID)
			}
		}
	}
}
