package pipedrive

import (
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	desc             = common.Descriptor("pipedrive", "Pipedrive", "/web/static/images/pipedrive.svg")
	companyDomainVar = sdktypes.NewSymbol("company_domain")
)
