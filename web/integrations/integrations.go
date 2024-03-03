package integrations

import (
	_ "embed"
	"html/template"
	"net/http"
	"net/url"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	defaultLogo = "" // TODO: Use some default logo.
)

//go:embed index.html
var index string

var tmpl = template.Must(template.New("index").Parse(index))

type Handler struct {
	fx.In

	Integrations sdkservices.Integrations
}

type tableRow struct {
	Name, Logo, Connection string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	is, err := h.Integrations.List(r.Context(), "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := []tableRow{}
	for _, i := range is {
		tr := tableRow{Name: i.DisplayName()}

		u := i.LogoURL()
		if u == nil {
			u = kittehs.Must1(url.Parse(defaultLogo))
		}
		tr.Logo = u.String()

		u = i.ConnectionURL()
		if u == nil {
			u = &url.URL{}
		}
		tr.Connection = u.String()

		data = append(data, tr)
	}

	if err := tmpl.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
