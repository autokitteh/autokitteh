package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
)

type deploymentStateTraits struct{}

var _ enumTraits = deploymentStateTraits{}

func (deploymentStateTraits) Prefix() string           { return "DEPLOYMENT_STATE_" }
func (deploymentStateTraits) Names() map[int32]string  { return deploymentsv1.DeploymentState_name }
func (deploymentStateTraits) Values() map[string]int32 { return deploymentsv1.DeploymentState_value }

type DeploymentState struct {
	enum[deploymentStateTraits, deploymentsv1.DeploymentState]
}

func deploymentStateFromProto(e deploymentsv1.DeploymentState) DeploymentState {
	return kittehs.Must1(DeploymentStateFromProto(e))
}

var (
	PossibleDeploymentStatesStrings = AllEnumStrings[deploymentStateTraits]()

	DeploymentStateUnspecified = deploymentStateFromProto(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_UNSPECIFIED)
	DeploymentStateActive      = deploymentStateFromProto(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_ACTIVE)
	DeploymentStateDraining    = deploymentStateFromProto(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_DRAINING)
	DeploymentStateInactive    = deploymentStateFromProto(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_INACTIVE)
	DeploymentStateTesting     = deploymentStateFromProto(deploymentsv1.DeploymentState_DEPLOYMENT_STATE_TESTING)
)

func DeploymentStateFromProto(e deploymentsv1.DeploymentState) (DeploymentState, error) {
	return EnumFromProto[DeploymentState](e)
}

func ParseDeploymentState(raw string) (DeploymentState, error) {
	return ParseEnum[DeploymentState](raw)
}
