package dashboardsvc

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (s *svc) initVars() {
	s.HandleFunc("DELETE "+rootPath+"vars/{sid}", s.deleteVars)
	s.HandleFunc("POST "+rootPath+"vars/{sid}", s.setVar)
}

func (s *svc) deleteVars(w http.ResponseWriter, r *http.Request) {
	keys := strings.Split(r.URL.Query()["names"][0], ",")
	syms, err := kittehs.TransformError(keys, sdktypes.StrictParseSymbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sid, err := sdktypes.Strict(sdktypes.ParseVarScopeID(r.PathValue("sid")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.Vars().Delete(r.Context(), sid, syms...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *svc) setVar(w http.ResponseWriter, r *http.Request) {
	sid, err := sdktypes.Strict(sdktypes.ParseVarScopeID(r.PathValue("sid")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Name       sdktypes.Symbol `json:"name"`
		Value      string          `json:"value"`
		IsSecret   bool            `json:"is_secret"`
		IsOptional bool            `json:"is_optional"`
	}

	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(bs, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v := sdktypes.NewVar(req.Name).SetValue(req.Value).SetSecret(req.IsSecret).WithScopeID(sid)

	if err := s.Vars().Set(r.Context(), v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *svc) genVarsList(w http.ResponseWriter, r *http.Request, sid sdktypes.VarScopeID) (list, error) {
	vs, err := s.Vars().Get(r.Context(), sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	cvars := list{
		Scope:   sid.String(),
		Headers: []string{"", "name", "value"},
		Items: kittehs.Transform(vs, func(cv sdktypes.Var) []template.HTML {
			v := cv.Value()
			if cv.IsSecret() {
				v = "🤐"
			}

			n := cv.Name().String()

			return []template.HTML{
				template.HTML(`<input type="checkbox" name="vars" value="` + n + `">`),
				template.HTML(n),
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
