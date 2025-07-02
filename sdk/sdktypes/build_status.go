package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	buildsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1"
)

type buildStatusTraits struct{}

var _ enumTraits = buildStatusTraits{}

func (buildStatusTraits) Prefix() string           { return "STATUS_" }
func (buildStatusTraits) Names() map[int32]string  { return buildsv1.Build_Status_name }
func (buildStatusTraits) Values() map[string]int32 { return buildsv1.Build_Status_value }

type BuildStatus struct {
	enum[buildStatusTraits, buildsv1.Build_Status]
}

func buildStatusFromProto(e buildsv1.Build_Status) BuildStatus {
	return kittehs.Must1(BuildStatusFromProto(e))
}

var (
	PossibleBuildStatusNames = AllEnumNames[buildStatusTraits]()

	BuildStatusUnspecified = buildStatusFromProto(buildsv1.Build_STATUS_UNSPECIFIED)
	BuildStatusPending     = buildStatusFromProto(buildsv1.Build_STATUS_PENDING)
	BuildStatusRunning     = buildStatusFromProto(buildsv1.Build_STATUS_RUNNING)
	BuildStatusSuccess     = buildStatusFromProto(buildsv1.Build_STATUS_SUCCESS)
	BuildStatusFailed      = buildStatusFromProto(buildsv1.Build_STATUS_FAILED)
)

func BuildStatusFromProto(e buildsv1.Build_Status) (BuildStatus, error) {
	return EnumFromProto[BuildStatus](e)
}

func ParseBuildStatus(raw string) (BuildStatus, error) {
	return ParseEnum[BuildStatus](raw)
}
