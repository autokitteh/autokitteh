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

func (s Svc) initVars() {
	s.Muxes.Auth.HandleFunc("DELETE /vars/{sid}", s.deleteVars)
	s.Muxes.Auth.HandleFunc("POST /vars/{sid}", s.setVar)
}

func (s Svc) deleteVars(w http.ResponseWriter, r *http.Request) {
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

	if err := s.Svcs.Vars().Delete(r.Context(), sid, syms...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s Svc) setVar(w http.ResponseWriter, r *http.Request) {
	sid, err := sdktypes.Strict(sdktypes.ParseVarScopeID(r.PathValue("sid")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req struct {
		Name       sdktypes.Symbol `json:"name"`
		Value      string          `json:"value"`
		IsSecret   bool            `json:"is_secret"`
		IsRequired bool            `json:"is_required"`
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

	v := sdktypes.NewVar(req.Name).SetValue(req.Value).SetSecret(req.IsSecret).SetRequired(req.IsRequired).WithScopeID(sid)

	if err := s.Svcs.Vars().Set(r.Context(), v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s Svc) genVarsList(w http.ResponseWriter, r *http.Request, sid sdktypes.VarScopeID) (list, error) {
	vs, err := s.Svcs.Vars().Get(r.Context(), sid)
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
			if cv.IsRequired() {
				n = "<em>" + n + "</em>"
			}

			return []template.HTML{
				template.HTML(`<input type="checkbox" name="vars" value="` + cv.Name().String() + `">`),
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
