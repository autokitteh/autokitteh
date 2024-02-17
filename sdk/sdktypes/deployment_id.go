package sdktypes

const DeploymentIDKind = "d"

type DeploymentID = *id[deploymentIDTraits]

var _ ID = (DeploymentID)(nil)

type deploymentIDTraits struct{}

func (deploymentIDTraits) Kind() string                   { return DeploymentIDKind }
func (deploymentIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseDeploymentID(raw string) (DeploymentID, error) {
	return parseTypedID[deploymentIDTraits](raw)
}

func StrictParseDeploymentID(raw string) (DeploymentID, error) {
	return strictParseTypedID[deploymentIDTraits](raw)
}

func NewDeploymentID() DeploymentID { return newID[deploymentIDTraits]() }
