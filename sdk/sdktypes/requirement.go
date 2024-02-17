package sdktypes

import (
	"fmt"
	"net/url"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

type RequirementPB = runtimesv1.Requirement

type Requirement = *object[*RequirementPB]

var (
	RequirementFromProto       = makeFromProto(validateRequirement)
	StrictRequirementFromProto = makeFromProto(strictValidateRequirement)
	ToStrictRequirement        = makeWithValidator(strictValidateRequirement)
)

func strictValidateRequirement(pb *runtimesv1.Requirement) error {
	return validateRequirement(pb)
}

func validateRequirement(pb *runtimesv1.Requirement) error {
	if _, err := CodeLocationFromProto(pb.Location); err != nil {
		return fmt.Errorf("location: %w", err)
	}

	if _, err := ParseSymbol(pb.Symbol); err != nil {
		return fmt.Errorf("symbol: %w", err)
	}

	if pb.Url != "" {
		if _, err := url.Parse(pb.Url); err != nil {
			return fmt.Errorf("url: %w", err)
		}
	}

	return nil
}

func NewRequirement(loc CodeLocation, url *url.URL, symbol Symbol) (Requirement, error) {
	return RequirementFromProto(&RequirementPB{
		Location: loc.ToProto(),
		Url:      url.String(),
		Symbol:   symbol.String(),
	})
}

func GetRequirementCodeLocation(r Requirement) CodeLocation {
	if r == nil {
		return nil
	}
	return kittehs.Must1(CodeLocationFromProto(r.pb.Location))
}

func GetRequirementURL(r Requirement) *url.URL {
	if r == nil {
		return nil
	}
	return kittehs.Must1(url.Parse(r.pb.Url))
}

func GetRequirementSymbol(r Requirement) Symbol {
	if r == nil {
		return nil
	}
	return kittehs.Must1(ParseSymbol(r.pb.Symbol))
}
