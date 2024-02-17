package apply

import (
	"context"

	"gopkg.in/yaml.v3"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type applicatorSvc struct {
	svcs sdkservices.Services
}

func NewApplySvc(svcs sdkservices.Services) sdkservices.Apply {
	return &applicatorSvc{
		svcs: svcs,
	}
}

func (s *applicatorSvc) Plan(ctx context.Context, manifest string) ([]string, error) {
	a, err := s.plan(ctx, manifest, "")
	if err != nil {
		return nil, err
	}

	return a.LogsAndOps(), nil
}

func (s *applicatorSvc) plan(ctx context.Context, manifest, path string) (*Applicator, error) {
	var root Root

	if err := yaml.Unmarshal([]byte(manifest), &root); err != nil {
		return nil, err
	}

	a := Applicator{
		Svcs: s.svcs,
		Path: path,
	}

	if err := a.Plan(ctx, &root); err != nil {
		return nil, err
	}

	return &a, nil
}

func (s *applicatorSvc) Apply(ctx context.Context, manifest, path string) ([]string, error) {
	a, err := s.plan(ctx, manifest, path)
	if err != nil {
		return nil, err
	}

	if err := a.Apply(ctx); err != nil {
		return nil, err
	}

	return a.LogsAndOps(), nil
}
