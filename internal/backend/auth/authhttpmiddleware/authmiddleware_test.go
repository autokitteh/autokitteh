package authhttpmiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens/authjwttokens"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authusers"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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

func assertCalledWithoutAuthn(t *testing.T, given *sdktypes.UserID) {
	if assert.NotNil(t, given) {
		assert.False(t, given.IsValid())
	}
}

// Returns a handler and a function that returns the authenticated user.
// If the handle was not called, the function will return nil.
// That function also resets the user and called states when its called.
func newInternalTestHandler(t *testing.T) (http.Handler, func() *sdktypes.UserID) {
	return newTestHandler(t, getCtxUserID)
}

func newOverallTestHandler(t *testing.T) (http.Handler, func() *sdktypes.UserID) {
	return newTestHandler(t, authcontext.GetAuthnUserID)
}

func newRequest(u sdktypes.User, authHeader string, cookies []*http.Cookie) *http.Request {
	req := kittehs.Must1(http.NewRequest(http.MethodGet, "/", nil))

	if u.IsValid() {
		req = req.WithContext(ctxWithUserID(context.Background(), u.ID()))
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	return req
}

func TestIfAuthenticated(t *testing.T) {
	yup, yupped := newInternalTestHandler(t)
	nope, noped := newInternalTestHandler(t)

	mw := ifAuthenticated(yup, nope)

	mw.ServeHTTP(nil, newRequest(sdktypes.InvalidUser, "", nil))
	assertCalledWithoutAuthn(t, noped())
	assert.Nil(t, yupped())

	mw.ServeHTTP(nil, newRequest(testUser1, "", nil))
	assert.Nil(t, noped())
	assertCalledWithUser(t, testUser1.ID(), yupped())
}

func TestDefaultUserMiddleware(t *testing.T) {
	h, check := newInternalTestHandler(t)

	mw := ifAuthenticated(h, newSetDefaultUserMiddleware(h))

	// When no other auth in context, should use default user.
	w := httptest.NewRecorder()
	mw.ServeHTTP(nil, newRequest(sdktypes.InvalidUser, "", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, authusers.DefaultUser.ID(), check())

	// When other auth in context, use that authn user.
	w = httptest.NewRecorder()
	mw.ServeHTTP(nil, newRequest(testUser1, "", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), check())
}

func TestTokensMiddleware(t *testing.T) {
	h, check := newInternalTestHandler(t)

	tokens := kittehs.Must1(authjwttokens.New(authjwttokens.Configs.Dev))

	mw := newTokensMiddleware(h, tokens)

	// no header - should be nop.
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithoutAuthn(t, check())

	// correct token.
	tok := kittehs.Must1(tokens.Create(testUser1))
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok, nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), check())

	// bad token.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "meow", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())
}

func TestSessionsMiddleware(t *testing.T) {
	h, check := newInternalTestHandler(t)

	sessions := kittehs.Must1(authsessions.New(authsessions.Configs.Dev))

	mw := newSessionsMiddleware(h, sessions)

	// no session - should be nop.
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithoutAuthn(t, check())

	w = httptest.NewRecorder()
	kittehs.Must0(sessions.Set(w, authsessions.NewSessionData(testUser1.ID())))
	cookies := w.Result().Cookies()

	// legit cookies.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", cookies))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), check())

	// bad cookie - nop.
	for i := range len(cookies) {
		w = httptest.NewRecorder()
		badCookies := cookies[:]
		badCookies[i].Value = "hiss"
		mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", badCookies))
		assert.Equal(t, http.StatusOK, w.Code)
		assertCalledWithoutAuthn(t, check())
	}
}

func TestNewWithoutDefaultUser(t *testing.T) {
	h, check := newOverallTestHandler(t)

	sessions := kittehs.Must1(authsessions.New(authsessions.Configs.Dev))
	tokens := kittehs.Must1(authjwttokens.New(authjwttokens.Configs.Dev))
	users := &sdktest.TestUsers{}

	mw := New(Deps{
		Sessions: sessions,
		Tokens:   tokens,
		Users:    users,
		Cfg:      &Config{UseDefaultUser: false},
	})(h)

	// no auth - reject.
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// correct token, but no such user.
	tok1 := kittehs.Must1(tokens.Create(testUser1))
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// correct token, but no such user.
	w = httptest.NewRecorder()
	kittehs.Must0(sessions.Set(w, authsessions.NewSessionData(testUser1.ID())))
	cookies := w.Result().Cookies()

	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// create a disabled user.
	users.Users = map[sdktypes.UserID]sdktypes.User{
		testUser1.ID(): testUser1.WithStatus(sdktypes.UserStatusDisabled),
	}

	// correct token, user is disabled.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// correct session, user is disabled.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// create a invited user.
	users.Users = map[sdktypes.UserID]sdktypes.User{
		testUser1.ID(): testUser1.WithStatus(sdktypes.UserStatusInvited),
	}

	// correct token, user is invited.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// correct session, user is invited.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// enable the user.
	users.Users = map[sdktypes.UserID]sdktypes.User{
		testUser1.ID(): testUser1,
	}

	// correct token, user is ok.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok1, nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), check())

	// correct session, user is ok.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", cookies))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), check())

	// both token and session for same user.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok1, cookies))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), check())

	// token for token user, session for other user.
	// token > session.
	users.Users[testUser2.ID()] = testUser2
	w = httptest.NewRecorder()
	tok2 := kittehs.Must1(tokens.Create(testUser2))
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok2, cookies))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser2.ID(), check())

	// token for token user who is disabled, session for other user.
	// token > session, and is rejected.
	users.Users[testUser2.ID()] = testUser2.WithStatus(sdktypes.UserStatusDisabled)
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok2, cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())
}

func TestNewWithDefaultUser(t *testing.T) {
	h, check := newOverallTestHandler(t)

	sessions := kittehs.Must1(authsessions.New(authsessions.Configs.Dev))
	tokens := kittehs.Must1(authjwttokens.New(authjwttokens.Configs.Dev))
	users := &sdktest.TestUsers{
		Users: map[sdktypes.UserID]sdktypes.User{
			authusers.DefaultUser.ID(): authusers.DefaultUser,
		},
	}

	mw := New(Deps{
		Sessions: sessions,
		Tokens:   tokens,
		Users:    users,
		Cfg:      &Config{UseDefaultUser: true},
	})(h)

	// no auth.
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, authusers.DefaultUser.ID(), check())

	// correct token, but no such user.
	tok1 := kittehs.Must1(tokens.Create(testUser1))
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// correct session, but no such user.
	w = httptest.NewRecorder()
	kittehs.Must0(sessions.Set(w, authsessions.NewSessionData(testUser1.ID())))
	cookies := w.Result().Cookies()

	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// create a disabled user.
	users.Users[testUser1.ID()] = testUser1.WithStatus(sdktypes.UserStatusDisabled)

	// correct token, user is disabled.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// correct token, user is disabled.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// create an invited user.
	users.Users[testUser1.ID()] = testUser1.WithStatus(sdktypes.UserStatusInvited)

	// correct token, user is invited.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok1, nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// correct token, user is invited.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", cookies))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, check())

	// enable the user.
	users.Users = map[sdktypes.UserID]sdktypes.User{
		testUser1.ID(): testUser1,
	}

	// correct token, user is ok.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "Bearer "+tok1, nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), check())

	// correct session, user is ok.
	w = httptest.NewRecorder()
	mw.ServeHTTP(w, newRequest(sdktypes.InvalidUser, "", cookies))
	assert.Equal(t, http.StatusOK, w.Code)
	assertCalledWithUser(t, testUser1.ID(), check())
}
