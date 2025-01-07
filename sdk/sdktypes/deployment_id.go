package sdktypes

const DeploymentIDKind = "dep"

type DeploymentID = id[deploymentIDTraits]

type deploymentIDTraits struct{}

func (deploymentIDTraits) Prefix() string { return DeploymentIDKind }

func NewDeploymentID() DeploymentID                    { return newID[DeploymentID]() }
func ParseDeploymentID(s string) (DeploymentID, error) { return ParseID[DeploymentID](s) }

var InvalidDeploymentID DeploymentID
