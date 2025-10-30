package pipedrive

import (
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const IntegrationName = "pipedrive"

var (
	desc             = common.Descriptor(IntegrationName, "Pipedrive", "/web/static/images/pipedrive.svg")
	companyDomainVar = sdktypes.NewSymbol("company_domain")
)
