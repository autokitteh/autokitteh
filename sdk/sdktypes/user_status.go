package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	usersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
)

type userStatusTraits struct{}

var _ enumTraits = userStatusTraits{}

func (userStatusTraits) Prefix() string           { return "USER_STATUS_" }
func (userStatusTraits) Names() map[int32]string  { return usersv1.UserStatus_name }
func (userStatusTraits) Values() map[string]int32 { return usersv1.UserStatus_value }

type UserStatusPB = usersv1.UserStatus

type UserStatus struct {
	enum[userStatusTraits, usersv1.UserStatus]
}

func userStatusFromProto(e usersv1.UserStatus) UserStatus {
	return kittehs.Must1(UserStatusFromProto(e))
}

var (
	PossibleUserStatusNames = AllEnumNames[userStatusTraits]()

	UserStatusUnspecified = userStatusFromProto(usersv1.UserStatus_USER_STATUS_UNSPECIFIED)
	UserStatusActive      = userStatusFromProto(usersv1.UserStatus_USER_STATUS_ACTIVE)
	UserStatusInvited     = userStatusFromProto(usersv1.UserStatus_USER_STATUS_INVITED)
	UserStatusDisabled    = userStatusFromProto(usersv1.UserStatus_USER_STATUS_DISABLED)
)

func UserStatusFromProto(e usersv1.UserStatus) (UserStatus, error) {
	return EnumFromProto[UserStatus](e)
}

func ParseUserStatus(raw string) (UserStatus, error) {
	return ParseEnum[UserStatus](raw)
}
