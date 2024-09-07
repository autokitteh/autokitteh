package sdktypes

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

const userIDKind = "usr"

type UserID = id[userIDTraits]

type userIDTraits struct{}

func (userIDTraits) Prefix() string { return userIDKind }

func ParseUserID(s string) (UserID, error)       { return ParseID[UserID](s) }
func StrictParseUserID(s string) (UserID, error) { return Strict(ParseUserID(s)) }

var InvalidUserID UserID

var BuiltinDefaultUserID = kittehs.Must1(ParseUserID("usr_3kthdf1t000000000000000000"))

// TODO: ENG-1112
func NewUserIDFromUserData(provider string, email string, name string) UserID {
	if provider == "ak" && email == "a@k" && name == "dflt" {
		return BuiltinDefaultUserID
	}

	combined := fmt.Sprintf("%s:%s:%s", provider, email, name)

	// Hash the combined string using SHA-256 -> 32 bytes
	hasher := sha256.New()
	hasher.Write([]byte(combined))
	hashBytes := hasher.Sum(nil)

	// uuid suffix should be base32 26-bytes long string.
	// 16 first bytes of hash -> base32 -> max suffix len 26 chars
	uuid := kittehs.Must1(uuid.FromBytes(hashBytes[:16]))
	id := NewIDFromUUID[UserID](&uuid)
	return id
}
