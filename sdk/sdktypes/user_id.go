package sdktypes

const userIDKind = "usr"

type UserID = id[userIDTraits]

type userIDTraits struct{}

func (userIDTraits) Prefix() string { return userIDKind }

func NewUserID() UserID                    { return newID[UserID]() }
func ParseUserID(s string) (UserID, error) { return ParseID[UserID](s) }

func IsUserID(s string) bool { return IsIDOf[userIDTraits](s) }

var InvalidUserID UserID
