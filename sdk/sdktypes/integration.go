package sdktypes

import (
	"errors"
	"net/url"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	integrationv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
)

type Integration struct {
	object[*IntegrationPB, IntegrationTraits]
}

var InvalidIntegration Integration

type IntegrationPB = integrationv1.Integration

type IntegrationTraits struct{}

func (IntegrationTraits) Validate(m *IntegrationPB) error {
	return errors.Join(
		idField[IntegrationID]("integration_id", m.IntegrationId),
		nameField("unique_name", m.UniqueName),
		objectField[Module]("module", m.Module),
		urlField("logo_url", m.LogoUrl),
		urlField("connection_url", m.ConnectionUrl),
	)
}

func (IntegrationTraits) StrictValidate(m *IntegrationPB) error {
	return errors.Join(
		mandatory("integration_id", m.IntegrationId),
		mandatory("unique_name", m.UniqueName),
	)
}

func IntegrationFromProto(m *IntegrationPB) (Integration, error) { return FromProto[Integration](m) }
func StrictIntegrationFromProto(m *IntegrationPB) (Integration, error) {
	return Strict(IntegrationFromProto(m))
}

func (p Integration) ID() IntegrationID {
	return kittehs.Must1(ParseIntegrationID(p.read().IntegrationId))
}
func (p Integration) DisplayName() string          { return p.read().DisplayName }
func (p Integration) UniqueName() Symbol           { return forceSymbol(p.read().UniqueName) }
func (p Integration) Description() string          { return p.read().Description }
func (p Integration) UserLinks() map[string]string { return p.read().UserLinks }

func (p Integration) LogoURL() *url.URL {
	if p.m == nil {
		return nil
	}
	return kittehs.Must1(url.Parse(p.m.LogoUrl))
}

func (p Integration) ConnectionURL() *url.URL {
	if p.m == nil {
		return nil
	}
	return kittehs.Must1(url.Parse(p.m.ConnectionUrl))
}

func (p Integration) UpdateModule(m Module) Integration {
	return Integration{p.forceUpdate(func(pb *IntegrationPB) { pb.Module = m.ToProto() })}
}

func (p Integration) WithDescription(s string) Integration {
	return Integration{p.forceUpdate(func(pb *IntegrationPB) { pb.Description = s })}
}

func (p Integration) WithUserLinks(links map[string]string) Integration {
	return Integration{p.forceUpdate(func(pb *IntegrationPB) { pb.UserLinks = links })}
}

func (p Integration) WithModule(m Module) Integration {
	return Integration{p.forceUpdate(func(pb *IntegrationPB) { pb.Module = m.ToProto() })}
}

func (p Integration) OwnerID() OwnerID {
	return kittehs.Must1(ParseOwnerID(p.read().OwnerId))
}
