package svc

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/fatih/color"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/version"
)

//go:embed banner.txt
var banner string

var bannerTemplate = template.Must(template.New("banner").Parse(banner))

var plainBanner, colorBanner string

type bannerConfig struct {
	Show bool `koanf:"show"`
}

var bannerConfigs = configset.Set[bannerConfig]{
	Default: &bannerConfig{},
	Dev:     &bannerConfig{Show: true},
}

func sprint(plain bool, cfg *bannerConfig, opts RunOptions, addr, wpAddr, wpVersion, temporalFrontendAddr, temporalUIAddr string) string {
	if !cfg.Show {
		return ""
	}

	fieldColor, eyeColor := fmt.Sprint, fmt.Sprint

	if !plain {
		fieldColor = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
		eyeColor = color.New(color.FgGreen).Add(color.Bold).SprintFunc()
	}

	var mode string
	if opts.Mode != "" {
		mode = "Mode:           " + fieldColor(opts.Mode) + " "
	}

	if temporalFrontendAddr != "" {
		temporalFrontendAddr = "Temporal:       " + fieldColor(temporalFrontendAddr) + " "
	}

	if temporalUIAddr != "" {
		temporalUIAddr = "Temporal UI:    " + fieldColor(temporalUIAddr) + " "
	}

	webAddr := "http://" + addr
	if wpAddr != "" {
		webAddr = "http://" + wpAddr
	}

	if wpVersion != "" {
		wpVersion = " v" + wpVersion + ":"
	} else {
		wpVersion = ":         "
	}

	var buf bytes.Buffer

	kittehs.Must0(bannerTemplate.Execute(&buf, struct {
		Version              string
		PID                  string
		Addr                 string
		WebPlatformAddr      string
		Eye                  string
		WebAddr              string
		WebVer               string
		Mode                 string
		Temporal, TemporalUI string
	}{
		Version:         fieldColor(version.Version),
		PID:             fieldColor(strconv.Itoa(os.Getpid())),
		Addr:            fieldColor(addr),
		WebPlatformAddr: fieldColor(wpAddr),
		WebAddr:         fieldColor(webAddr),
		WebVer:          wpVersion,
		Eye:             eyeColor("▀"),
		Mode:            mode,
		Temporal:        temporalFrontendAddr,
		TemporalUI:      temporalUIAddr,
	}))

	return buf.String()
}

func initBanner(cfg *bannerConfig, opts RunOptions, httpsvcAddr, wpAddr, wpVersion, temporalFrontendAddr, temporalUIAddr string) {
	plainBanner = sprint(true, cfg, opts, httpsvcAddr, wpAddr, wpVersion, temporalFrontendAddr, temporalUIAddr)
	colorBanner = sprint(false, cfg, opts, httpsvcAddr, wpAddr, wpVersion, temporalFrontendAddr, temporalUIAddr)
}
