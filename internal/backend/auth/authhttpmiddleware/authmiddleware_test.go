package authhttpmiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens/authjwttokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktest"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	testUser1 = sdktypes.NewUser().WithNewID().WithStatus(sdktypes.UserStatusActive).WithDisplayName("Test User 1")
	testUser2 = sdktypes.NewUser().WithNewID().WithStatus(sdktypes.UserStatusActive).WithDisplayName("Test User 2")
)

func newTestHandler(t *testing.T, getUser func(context.Context) sdktypes.UserID) (http.Handler, func() *sdktypes.UserID) {
	var (
		uid    sdktypes.UserID
		called bool
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called, uid = true, getUser(r.Context())
			t.Logf("handler called, user_id=%v", uid)
		}),
		func() *sdktypes.UserID {
			u, c := uid, called
			uid, called = sdktypes.InvalidUserID, false

			if !c {
				return nil
			}

			return &u
		}
}

func assertCalledWithUser(t *testing.T, expected sdktypes.UserID, given *sdktypes.UserID) {
	if assert.NotNil(t, given) {
		assert.Equal(t, expected, *given)
	}
}

func newRequest(authHeader string, cookies []*http.Cookie) *http.Request {
	req := kittehs.Must1(http.NewRequest(http.MethodGet, "/", nil))

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	return req
}

func TestDefaultUserMiddleware(t *testing.T) {
	uid, mwErr := setDefaultUserMiddleware(newRequest("", nil))
	if assert.Nil(t, mwErr) {
		assert.Equal(t, authusers.DefaultUser.ID(), uid)
	}
}

func TestTokensMiddleware(t *testing.T) {
	tokens := kittehs.Must1(authjwttokens.New(authjwttokens.Configs.Dev))

	mw := newTokensMiddleware(tokens)

	// no header - should be nop.
	uid, mwErr := mw(newRequest("", nil))
	if assert.Nil(t, mwErr) {
		assert.False(t, uid.IsValid())
	}

	// correct token.
	tok := kittehs.Must1(tokens.Create(testUser1))
	uid, mwErr = mw(newRequest("Bearer "+tok, nil))
	if assert.Nil(t, mwErr) {
		assert.True(t, uid.IsValid())
	}

	// bad token.
	uid, mwErr = mw(newRequest("hiss", nil))
	if assert.NotNil(t, mwErr) {
		assert.Equal(t, http.StatusUnauthorized, mwErr.code)
	}
	assert.False(t, uid.IsValid())
}

func TestSessionsMiddleware(t *testing.T) {
	sessions := kittehs.Must1(authsessions.New(authsessions.Configs.Dev))

	mw := newSessionsMiddleware(sessions)

	// no session - should be nop.
	uid, mwErr := mw(newRequest("", nil))
	if assert.Nil(t, mwErr) {
		assert.False(t, uid.IsValid())
	}

	w := httptest.NewRecorder()
	kittehs.Must0(sessions.Set(w, authsessions.NewSessionData(testUser1.ID())))
	cookies := w.Result().Cookies()

	// legit cookies.
	uid, mwErr = mw(newRequest("", cookies))
	if assert.Nil(t, mwErr) {
		assert.True(t, uid.IsValid())
	}

	// bad cookie - nop (a bad cookie should not fail the middleware).
	for i := range cookies {
		badCookies := make([]*http.Cookie, len(cookies))
		copy(badCookies, cookies)
		badCookies[i].Value = "hiss"

		uid, mwErr = mw(newRequest("", badCookies))
		if assert.Nil(t, mwErr) {
			assert.False(t, uid.IsValid())
		}
	}
}

type harness struct {
	mw       http.Handler
	check    func() *sdktypes.UserID
	sessions authsessions.Store
	tokens   authtokens.Tokens
	users    *sdktest.TestUsers
}

type testUsers struct{ sdkservices.Users }

func (u testUsers) HasDefaultUser() bool {
	_, err := u.Get(context.TODO(), authusers.DefaultUser.ID(), "")
	return err == nil
}

func (u testUsers) Setup(context.Context) error { return nil }

func newTestHarness(t *testing.T, useDefaultUser bool) *harness {
	h, check := newTestHandler(t, authcontext.GetAuthnUserID)

	sessions := kittehs.Must1(authsessions.New(authsessions.Configs.Dev))
	tokens := kittehs.Must1(authjwttokens.New(authjwttokens.Configs.Dev))
	users := &sdktest.TestUsers{}

	if useDefaultUser {
		users.Users = map[sdktypes.UserID]sdktypes.User{
			authusers.DefaultUser.ID(): authusers.DefaultUser,
		}
	}

	mw := New(Deps{
		Logger:   zaptest.NewLogger(t),
		Sessions: sessions,
		Tokens:   tokens,
		Users:    testUsers{users},
	})(h)

	return &harness{
		mw:       mw,
		check:    check,
		sessions: sessions,
		tokens:   tokens,
		users:    users,
	}
}

func TestMiddlewareError(t *testing.T) {
	h := newTestHarness(t, false)

	w := httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer hisss", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	assert.Equal(t, "invalid token\n", w.Body.String())
}

func TestNewWithoutDefaultUser(t *testing.T) {
	h := newTestHarness(t, false)

	// no auth - reject.
	w := httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// correct token, but no such user.
	tok1 := kittehs.Must1(h.tokens.Create(testUser1))
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// correct token, but no such user.
	w = httptest.NewRecorder()
	kittehs.Must0(h.sessions.Set(w, authsessions.NewSessionData(testUser1.ID())))
	cookies := w.Result().Cookies()

	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// create a disabled user.
	h.users.Users = map[sdktypes.UserID]sdktypes.User{
		testUser1.ID(): testUser1.WithStatus(sdktypes.UserStatusDisabled),
	}

	// correct token, user is disabled.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// correct session, user is disabled.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// create a invited user.
	h.users.Users = map[sdktypes.UserID]sdktypes.User{
		testUser1.ID(): testUser1.WithStatus(sdktypes.UserStatusInvited),
	}

	// correct token, user is invited.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// correct session, user is invited.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// enable the user.
	h.users.Users = map[sdktypes.UserID]sdktypes.User{
		testUser1.ID(): testUser1,
	}

	// correct token, user is ok.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok1, nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), h.check())

	// correct session, user is ok.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", cookies))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), h.check())

	// both token and session for same user.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok1, cookies))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), h.check())

	// token for token user, session for other user.
	// token > session.
	h.users.Users[testUser2.ID()] = testUser2
	w = httptest.NewRecorder()
	tok2 := kittehs.Must1(h.tokens.Create(testUser2))
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok2, cookies))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser2.ID(), h.check())

	// token for token user who is disabled, session for other user.
	// token > session, and is rejected.
	h.users.Users[testUser2.ID()] = testUser2.WithStatus(sdktypes.UserStatusDisabled)
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok2, cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())
}

func TestNewWithDefaultUser(t *testing.T) {
	h := newTestHarness(t, true)

	// no auth.
	w := httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, authusers.DefaultUser.ID(), h.check())

	// correct token, but no such user.
	tok1 := kittehs.Must1(h.tokens.Create(testUser1))
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// correct session, but no such user.
	w = httptest.NewRecorder()
	kittehs.Must0(h.sessions.Set(w, authsessions.NewSessionData(testUser1.ID())))
	cookies := w.Result().Cookies()

	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// create a disabled user.
	h.users.Users[testUser1.ID()] = testUser1.WithStatus(sdktypes.UserStatusDisabled)

	// correct token, user is disabled.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// correct token, user is disabled.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// create an invited user.
	h.users.Users[testUser1.ID()] = testUser1.WithStatus(sdktypes.UserStatusInvited)

	// correct token, user is invited.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// correct token, user is invited.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, h.check())

	// enable the user.
	h.users.Users = map[sdktypes.UserID]sdktypes.User{
		testUser1.ID(): testUser1,
	}

	// correct token, user is ok.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("Bearer "+tok1, nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), h.check())

	// correct session, user is ok.
	w = httptest.NewRecorder()
	h.mw.ServeHTTP(w, newRequest("", cookies))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), h.check())
}
