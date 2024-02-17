package sdktypes

import (
	"fmt"
	"net/url"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
)

type IntegrationPB = integrationsv1.Integration

type Integration = *object[*IntegrationPB]

var (
	IntegrationFromProto       = makeFromProto(validateIntegration)
	StrictIntegrationFromProto = makeFromProto(strictValidateIntegration)
	ToStrictIntegration        = makeWithValidator(strictValidateIntegration)
)

func strictValidateIntegration(pb *integrationsv1.Integration) error {
	if err := ensureNotEmpty(pb.IntegrationId, pb.UniqueName); err != nil {
		return err
	}
	return validateIntegration(pb)
}

func validateIntegration(pb *integrationsv1.Integration) error {
	// Required fields.
	if _, err := ParseIntegrationID(pb.IntegrationId); err != nil {
		return fmt.Errorf("integration ID: %w", err)
	}

	if _, err := ParseName(pb.UniqueName); err != nil {
		return fmt.Errorf("integration's unique name: %w", err)
	}

	// TODO(ENG-346): Connection UI specification instead of a URL.
	if _, err := url.Parse(pb.ConnectionUrl); err != nil {
		return fmt.Errorf("integration's connection URL: %w", err)
	}

	// TODO: Methods.

	// TODO: Events.

	// Optional fields (must be valid if present).
	if _, err := url.Parse(pb.LogoUrl); err != nil {
		return fmt.Errorf("integration's logo URL: %w", err)
	}

	for _, v := range pb.UserLinks {
		if _, err := url.Parse(v); err != nil {
			return fmt.Errorf("integration user link %q: %w", v, err)
		}
	}

	return nil
}

func GetIntegrationID(i Integration) IntegrationID {
	if i == nil {
		return nil
	}

	return kittehs.Must1(ParseIntegrationID(i.pb.IntegrationId))
}

func GetIntegrationUniqueName(i Integration) Name {
	if i == nil {
		return nil
	}

	return kittehs.Must1(ParseName(i.pb.UniqueName))
}

func GetIntegrationDisplayName(i Integration) string {
	if i == nil {
		return ""
	}

	if i.pb.DisplayName != "" {
		return i.pb.DisplayName
	}
	return i.pb.UniqueName
}

func GetIntegrationDescription(i Integration) string {
	if i == nil {
		return ""
	}

	return i.pb.Description
}

func GetIntegrationLogoURL(i Integration) *url.URL {
	if i == nil {
		return nil
	}

	u, err := url.ParseRequestURI(i.pb.LogoUrl)
	if err != nil {
		return nil
	}
	return u
}

func GetIntegrationUserLinks(i Integration) map[string]*url.URL {
	if i == nil {
		return nil
	}

	m := map[string]*url.URL{}
	for k, v := range i.pb.UserLinks {
		if u, err := url.ParseRequestURI(v); err == nil {
			m[k] = u
		}
	}
	return m
}

// TODO: func GetIntegrationTag/s

// TODO(ENG-346): Connection UI specification instead of a URL.
func GetIntegrationConnectionURL(i Integration) *url.URL {
	if i == nil {
		return nil
	}

	u, err := url.ParseRequestURI(i.pb.ConnectionUrl)
	if err != nil {
		return nil
	}
	return u
}

func GetIntegrationModule(i Integration) Module {
	if i == nil {
		return nil
	}

	return kittehs.Must1(ModuleFromProto(i.pb.Module))
}

// TODO: Methods.
// TODO: Events.
