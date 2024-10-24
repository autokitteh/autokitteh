package dashboardsvc

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	//                                     +-- if it's part of a path, we don't want it.
	//                                     V
	idRegexp       = regexp.MustCompile(`([^/])([a-z]{3}_[` + sdktypes.ValidIDChars + `]{26})`)
	urlRegexp      = regexp.MustCompile(`https?://[^\s"]+`)
	urlFieldRegexp = regexp.MustCompile(`_url":[\s]*"([^"]+)"`)
)

func formatField(k, v string) template.HTML {
	if strings.HasSuffix(k, "_id") && sdktypes.IsID(v) {
		v = fmt.Sprintf("<a href='"+rootPath+"objects/%s'>%s</a>", v, v)
	} else if strings.HasSuffix(k, "_url") {
		v = fmt.Sprintf("<a href='%s'>%s</a>", v, v)
	}

	return template.HTML(v)
}

func formatText(txt string) template.HTML {
	txt = idRegexp.ReplaceAllString(txt, `$1<a href="`+rootPath+`objects/$2">$2</a>`)
	txt = urlRegexp.ReplaceAllString(txt, `<a href="$0">$0</a>`)
	txt = urlFieldRegexp.ReplaceAllString(txt, `_url": "<a href="$1">$1</a>"`)
	return template.HTML(txt)
}
