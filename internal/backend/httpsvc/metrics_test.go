package httpsvc

import (
	"testing"
)

func TestGetMetricNameFromPath(t *testing.T) {
	tests := []struct {
		name, in, out string
	}{
		{
			"irrelevant_path", "irrelevant_path", "",
		},
		{
			"v1-projects",
			"/autokitteh.projects.v1.ProjectsService/Create",
			"projects.v1.create",
		},
		{
			"v2-projects",
			"/autokitteh.projects.v2.ProjectsService/Create",
			"projects.v2.create",
		},
		{
			"v2-other-projects",
			"/autokitteh.projects.v2.OtherProjectsService/Create",
			"projects_otherprojects.v2.create",
		},
		{
			"v1-other-projects",
			"/autokitteh.projects.v1.OtherProjectsService/Create",
			"projects_otherprojects.v1.create",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMetricNameFromPath(tt.in); got != tt.out {
				t.Errorf("got %q, want %q", got, tt)
			}
		})
	}
}
