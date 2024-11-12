package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
)

type Deployment struct {
	object[*DeploymentPB, DeploymentTraits]
}

var InvalidDeployment Deployment

type DeploymentPB = deploymentv1.Deployment

type DeploymentTraits struct{}

func (DeploymentTraits) Validate(m *DeploymentPB) error {
	return errors.Join(
		idField[DeploymentID]("deployment_id", m.DeploymentId),
		idField[ProjectID]("project_id", m.ProjectId),
		idField[BuildID]("build_id", m.BuildId),
		enumField[DeploymentState]("state", m.State),
	)
}

func (DeploymentTraits) StrictValidate(m *DeploymentPB) error {
	return errors.Join(
		mandatory("deployment_id", m.DeploymentId),
		mandatory("project_id", m.ProjectId),
		mandatory("build_id", m.BuildId),
	)
}

func DeploymentFromProto(m *DeploymentPB) (Deployment, error) { return FromProto[Deployment](m) }
func StrictDeploymentFromProto(m *DeploymentPB) (Deployment, error) {
	return Strict(DeploymentFromProto(m))
}

func NewDeployment(id DeploymentID, envID ProjectID, buildID BuildID) Deployment {
	return kittehs.Must1(DeploymentFromProto(&DeploymentPB{DeploymentId: id.String(), ProjectId: envID.String(), BuildId: buildID.String()}))
}

func (p Deployment) ID() DeploymentID { return kittehs.Must1(ParseDeploymentID(p.read().DeploymentId)) }

func (p Deployment) WithNewID() Deployment {
	return Deployment{p.forceUpdate(func(pb *DeploymentPB) { pb.DeploymentId = NewDeploymentID().String() })}
}

func (p Deployment) WithID(id DeploymentID) Deployment {
	return Deployment{p.forceUpdate(func(pb *DeploymentPB) { pb.DeploymentId = id.String() })}
}

func (p Deployment) ProjectID() ProjectID { return kittehs.Must1(ParseProjectID(p.read().ProjectId)) }
func (p Deployment) BuildID() BuildID     { return kittehs.Must1(ParseBuildID(p.read().BuildId)) }
func (p Deployment) WithoutTimestamps() Deployment {
	return Deployment{p.forceUpdate(func(pb *DeploymentPB) {
		pb.CreatedAt = nil
		pb.UpdatedAt = nil
	})}
}

func (p Deployment) State() DeploymentState {
	return forceEnumFromProto[DeploymentState](p.read().State)
}

func (p Deployment) WithState(s DeploymentState) Deployment {
	return Deployment{p.forceUpdate(func(pb *DeploymentPB) { pb.State = s.ToProto() })}
}
