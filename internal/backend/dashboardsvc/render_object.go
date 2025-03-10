package dashboardsvc

import (
	_ "embed"
	"html/template"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

var marshalOpts = protojson.MarshalOptions{
	Multiline:     true,
	Indent:        "  ",
	UseProtoNames: true,
}

func marshalObject(x proto.Message) template.HTML {
	return formatText(string(kittehs.Must1(marshalOpts.Marshal(x))))
}

func renderObject[M proto.Message](w http.ResponseWriter, r *http.Request, title string, x M) {
	n := "object.html"
	json := marshalObject(x)

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, n, struct {
		Message string
		Title   string
		JSON    template.HTML
		Extra   any
	}{
		Title: title,
		JSON:  json,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
