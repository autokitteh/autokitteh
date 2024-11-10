package sdktypes

import "go.autokitteh.dev/autokitteh/internal/kittehs"

const userIDKind = "usr"

type UserID = id[userIDTraits]

type userIDTraits struct{}

func (userIDTraits) Prefix() string { return userIDKind }

func ParseUserID(s string) (UserID, error)       { return ParseID[UserID](s) }
func StrictParseUserID(s string) (UserID, error) { return Strict(ParseUserID(s)) }

var InvalidUserID UserID

func NewUserID() UserID      { return newID[UserID]() }
func IsUserID(s string) bool { return IsIDOf[userIDTraits](s) }

func NewTestUserID(email string) UserID {
	return kittehs.Must1(ParseUserID(newNamedIDString(email, userIDKind)))
}
