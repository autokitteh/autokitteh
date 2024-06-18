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

// func NewUserID() UserID                          { return newID[UserID]() }
func ParseUserID(s string) (UserID, error)       { return ParseID[UserID](s) }
func StrictParseUserID(s string) (UserID, error) { return Strict(ParseUserID(s)) }

var InvalidUserID UserID

// FIXME: use provider ID
func NewUserIDFromUserData(provider string, email string, name string) UserID {
	// siffix should be base32 26-bytes long string
	pLen := len(provider)
	if pLen > 3 {
		pLen = 3
	}
	combined := fmt.Sprintf("%03s:%s:%s", provider[:pLen], email, name)

	// Hash the combined string using SHA-256 -> 32 bytes
	hasher := sha256.New()
	hasher.Write([]byte(combined))
	hashBytes := hasher.Sum(nil)

	// 16 first bytes of hash -> base32 -> max suffix len 26 chars
	// encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hashBytes[:16])

	// id := fmt.Sprintf("%s_%c%s", userIDKind, rune('2'+(int(encoded[0])%6)), encoded[1:26])

	// typeid requires that first char should be in 0-7 range, but 0-1 are illegal in base32

	// fmt.Println("id:", id)
	// return kittehs.Must1(ParseUserID(id))

	uuid := kittehs.Must1(uuid.FromBytes(hashBytes[:16]))
	id := NewIDFromUUID[UserID](&uuid)
	return id

	// typeid.FromUUIDBytes[]]("usr", hashBytes[:16])
}
