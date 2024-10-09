package svc

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/fatih/color"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
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
var banner string

var bannerTemplate = template.Must(template.New("banner").Parse(banner))

func printBanner(cfg *bannerConfig, opts RunOptions, addr, wpAddr, temporalFrontendAddr, temporalUIAddr string) {
	if !cfg.Show {
		return
	}

	fieldColor := color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	eyeColor := color.New(color.FgGreen).Add(color.Bold).SprintFunc()

	var mode string
	if opts.Mode != "" {
		mode = "Mode:        " + fieldColor(opts.Mode) + " "
	}

	if temporalFrontendAddr != "" {
		temporalFrontendAddr = "Temporal:    " + fieldColor(temporalFrontendAddr) + " "
	}

	if temporalUIAddr != "" {
		temporalUIAddr = "Temporal UI: " + fieldColor(temporalUIAddr) + " "
	}

	webAddr := fmt.Sprintf("http://%s", addr)
	if wpAddr != "" {
		webAddr = fmt.Sprintf("http://%s", wpAddr)
	}

	kittehs.Must0(bannerTemplate.Execute(os.Stderr, struct {
		Version              string
		PID                  string
		Addr                 string
		WebPlatformAddr      string
		Eye                  string
		WebAddr              string
		Mode                 string
		Temporal, TemporalUI string
	}{
		Version:         fieldColor(version.Version),
		PID:             fieldColor(fmt.Sprintf("%d", os.Getpid())),
		Addr:            fieldColor(addr),
		WebPlatformAddr: fieldColor(wpAddr),
		WebAddr:         fieldColor(webAddr),
		Eye:             eyeColor("▀"),
		Mode:            mode,
		Temporal:        temporalFrontendAddr,
		TemporalUI:      temporalUIAddr,
	}))
}
