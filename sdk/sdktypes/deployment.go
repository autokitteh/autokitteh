package sdktypes

import (
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type (
	DeploymentPB = deploymentsv1.Deployment
	Deployment   = *object[*DeploymentPB]
)

type DeploymentState deploymentsv1.DeploymentState

type DeploymentSessionsStats = deploymentsv1.Deployment_SessionStats

const (
	DeploymentStateUnspecified = DeploymentState(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_UNSPECIFIED)
	DeploymentStateActive      = DeploymentState(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_ACTIVE)
	DeploymentStateDraining    = DeploymentState(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_DRAINING)
	DeploymentStateInactive    = DeploymentState(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_INACTIVE)
	DeploymentStateTesting     = DeploymentState(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_TESTING)
)

func DeploymentStateFromProto(s deploymentsv1.DeploymentState) (DeploymentState, error) {
	if _, ok := deploymentsv1.DeploymentState_name[int32(s.Number())]; ok {
		return DeploymentState(s), nil
	}
	return DeploymentStateUnspecified, fmt.Errorf("unknown state %v: %w", s, sdkerrors.ErrInvalidArgument)
}

func (s DeploymentState) String() string {
	return strings.TrimPrefix(deploymentsv1.DeploymentState_name[int32(s)], "DEPLOYMENT_STATE_")
}

func (s DeploymentState) ToProto() deploymentsv1.DeploymentState {
	return deploymentsv1.DeploymentState(s)
}

func ParseDeploymentState(raw string) DeploymentState {
	if raw == "" {
		return DeploymentStateUnspecified
	}
	upper := strings.ToUpper(raw)
	if !strings.HasPrefix(upper, "DEPLOYMENT_STATE_") {
		upper = "DEPLOYMENT_STATE_" + upper
	}

	state, ok := deploymentsv1.DeploymentState_value[upper]
	if !ok {
		return DeploymentStateUnspecified
	}

	return DeploymentState(state)
}

var (
	DeploymentFromProto       = makeFromProto(validateDeployment)
	StrictDeploymentFromProto = makeFromProto(strictValidateDeployment)
	ToStrictDeployment        = makeWithValidator(strictValidateDeployment)
)

func strictValidateDeployment(pb *deploymentsv1.Deployment) error {
	if err := ensureNotEmpty(pb.DeploymentId, pb.EnvId, pb.BuildId); err != nil {
		return err
	}

	return validateDeployment(pb)
}

func validateDeployment(pb *deploymentsv1.Deployment) error {
	if _, err := ParseDeploymentID(pb.DeploymentId); err != nil {
		return fmt.Errorf("Deployment id: %w", err)
	}

	if _, err := ParseEnvID(pb.EnvId); err != nil {
		return fmt.Errorf("env id: %w", err)
	}

	if _, err := ParseBuildID(pb.BuildId); err != nil {
		return fmt.Errorf("event id: %w", err)
	}

	return nil
}

func GetDeploymentID(e Deployment) DeploymentID {
	if e == nil {
		return nil
	}
	return kittehs.Must1(ParseDeploymentID(e.pb.DeploymentId))
}

func GetDeploymentEnvID(e Deployment) EnvID {
	if e == nil {
		return nil
	}
	return kittehs.Must1(ParseEnvID(e.pb.EnvId))
}

func GetDeploymentBuildID(e Deployment) BuildID {
	if e == nil {
		return nil
	}
	return kittehs.Must1(ParseBuildID(e.pb.BuildId))
}

func GetDeploymentState(e Deployment) DeploymentState {
	return DeploymentState(e.pb.State)
}

var PossibleDeploymentStates = kittehs.Transform(kittehs.MapValuesSortedByKeys(deploymentsv1.DeploymentState_name), func(name string) string {
	return strings.TrimPrefix(name, "DEPLOYMENT_STATE_")
})

func DeploymentWithoutTimes(d Deployment) Deployment {
	if d == nil {
		return nil
	}

	return kittehs.Must1(d.Update(func(pb *DeploymentPB) {
		pb.CreatedAt = nil
		pb.UpdatedAt = nil
	}))
}
