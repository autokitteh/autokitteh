package dashboardsvc

import (
	"html/template"
	"net/http"
	"sort"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (s Svc) genVarsList(w http.ResponseWriter, r *http.Request, sid sdktypes.VarScopeID) (list, error) {
	vs, err := s.Svcs.Vars().Get(r.Context(), sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	cvars := list{
		Headers: []string{"name", "value"},
		Items: kittehs.Transform(vs, func(cv sdktypes.Var) []template.HTML {
			v := cv.Value()
			if cv.IsSecret() {
				v = "🤐"
			}
			return []template.HTML{
				template.HTML(cv.Name().String()),
				template.HTML(v),
			}
		}),
		N: len(vs),
	}

	sort.Slice(cvars.Items, func(i, j int) bool {
		return cvars.Items[i][0] < cvars.Items[j][0]
	})

	return cvars, nil
}
