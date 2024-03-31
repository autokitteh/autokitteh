package sdktypes

const deploymentIDKind = "dep"

type DeploymentID = id[deploymentIDTraits]

type deploymentIDTraits struct{}

func (deploymentIDTraits) Prefix() string { return deploymentIDKind }

func NewDeploymentID() DeploymentID                    { return newID[DeploymentID]() }
func ParseDeploymentID(s string) (DeploymentID, error) { return ParseID[DeploymentID](s) }

var InvalidDeploymentID DeploymentID
