package authloginhttpsvc

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func newLegacyUserIDFromUserData(ld *loginData) sdktypes.UserID {
	combined := fmt.Sprintf("%s:%s:%s", ld.ProviderName, ld.Email, ld.DisplayName)

	// Hash the combined string using SHA-256 -> 32 bytes
	hasher := sha256.New()
	hasher.Write([]byte(combined))
	hashBytes := hasher.Sum(nil)

	// uuid suffix should be base32 26-bytes long string.
	// 16 first bytes of hash -> base32 -> max suffix len 26 chars
	uuid := kittehs.Must1(uuid.FromBytes(hashBytes[:16]))
	return sdktypes.NewIDFromUUID[sdktypes.UserID](&uuid)
}
