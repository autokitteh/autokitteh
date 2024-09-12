package httpsvc

import "testing"

func TestGetMetricNameFromPath(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{
			"irrelevant_path", "",
		},
		{
			"/autokitteh.projects.v1.ProjectsService/Create", "projects.v1.create",
		},
		{
			"/autokitteh.projects.v2.ProjectsService/Create", "projects.v2.create",
		},
		{
			"/autokitteh.projects.v2.OtherProjectsService/Create", "projects_otherprojects.v2.create",
		},
		{
			"/autokitteh.projects.v1.OtherProjectsService/Create", "projects_otherprojects.v1.create",
		},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			if got := getMetricNameFromPath(test.in); got != test.out {
				t.Errorf("got %q, want %q", got, test)
			}
		})
	}
}
