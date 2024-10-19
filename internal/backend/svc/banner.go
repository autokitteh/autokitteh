package svc

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/backend/webplatform"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/version"
)

type bannerConfig struct {
	Show bool `koanf:"show"`
}

var bannerConfigs = configset.Set[bannerConfig]{
	Default: &bannerConfig{},
	Dev:     &bannerConfig{Show: true},
}

//go:embed banner.txt
var bannerTxt string

var bannerTemplate = template.Must(template.New("banner").Parse(bannerTxt))

type banner struct {
	fx.In

	HTTPSvc httpsvc.Svc
	TClient temporalclient.Client
	Web     *webplatform.Svc
}

func (b *banner) Print(w io.Writer, opts RunOptions, html bool) {
	temporalFrontendAddr, temporalUIAddr := b.TClient.TemporalAddr()

	var (
		fielder func(...any) string
		eye     = "▀"
	)

	if html {
		fielder = func(a ...interface{}) string {
			v := fmt.Sprint(a...)

			if strings.HasPrefix(v, "http:") || strings.HasPrefix(v, "https:") {
				v = fmt.Sprintf(`<a style="color: inherit;" href="%s">%s</a>`, v, v)
			}

			v = fmt.Sprintf("<strong><span style=\"color:rgb(0, 150, 255)\">%s</span></strong>", v)

			return v
		}
		eye = `<strong><span style="color:rgb(127, 255, 212)">` + eye + `</span></strong>`
	} else {
		fielder = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
		eye = color.New(color.FgGreen).Add(color.Bold).SprintFunc()(eye)
	}

	var mode string
	if opts.Mode != "" {
		mode = "Mode:         " + fielder(opts.Mode) + " "
	}

	if temporalFrontendAddr != "" {
		temporalFrontendAddr = "Temporal:     " + fielder(temporalFrontendAddr) + " "
	}

	if temporalUIAddr != "" {
		temporalUIAddr = "Temporal UI:  " + fielder(temporalUIAddr) + " "
	}

	addr := b.HTTPSvc.MainAddr()
	wpAddr := b.Web.Addr()
	wpVersion := b.Web.Version()

	webAddr := fmt.Sprintf("http://%s", addr)
	if wpAddr != "" {
		webAddr = fmt.Sprintf("http://%s", wpAddr)
	}

	auxAddr := fmt.Sprintf("http://%s", b.HTTPSvc.AuxAddr())

	if wpVersion != "" {
		wpVersion = " v" + wpVersion + ":"
	} else {
		wpVersion = ":        "
	}

	kittehs.Must0(bannerTemplate.Execute(w, struct {
		Version              string
		PID                  string
		Addr                 string
		WebPlatformAddr      string
		Eye                  string
		WebAddr              string
		WebVer               string
		AuxAddr              string
		Mode                 string
		Temporal, TemporalUI string
	}{
		Version:         fielder(version.Version),
		PID:             fielder(fmt.Sprintf("%d", os.Getpid())),
		Addr:            fielder(addr),
		WebPlatformAddr: fielder(wpAddr),
		WebAddr:         fielder(webAddr),
		WebVer:          wpVersion,
		AuxAddr:         fielder(auxAddr),
		Eye:             eye,
		Mode:            mode,
		Temporal:        temporalFrontendAddr,
		TemporalUI:      temporalUIAddr,
	}))
}
