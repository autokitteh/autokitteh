package httpsvc

import (
	"strconv"
	"testing"
)

func TestGetMetricNameFromPath(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{
			"irrelevant_path", "",
		},
		{
			"/autokitteh.projects.v1.ProjectsService/Create",
			"projects.v1.create",
		},
		{
			"/autokitteh.projects.v2.ProjectsService/Create",
			"projects.v2.create",
		},
		{
			"/autokitteh.projects.v2.OtherProjectsService/Create",
			"projects_otherprojects.v2.create",
		},
		{
			"/autokitteh.projects.v1.OtherProjectsService/Create",
			"projects_otherprojects.v1.create",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := getMetricNameFromPath(tt.in); got != tt.out {
				t.Errorf("got %q, want %q", got, tt)
			}
		})
	}
}
