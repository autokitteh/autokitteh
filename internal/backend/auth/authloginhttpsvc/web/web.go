package web

import (
	_ "embed"
	"text/template"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed login.html
var loginHTML string

var LoginTemplate = kittehs.Must1(template.New("").Parse(loginHTML))

//go:embed descope.html
var descopeLoginHTML string

var DescopeLoginTemplate = kittehs.Must1(template.New("").Parse(descopeLoginHTML))
