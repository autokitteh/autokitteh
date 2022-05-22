package svc

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/ghodss/yaml"
	"github.com/kelseyhightower/envconfig"

	"github.com/autokitteh/L"
	"github.com/autokitteh/L/Z"
)

type httpCfg struct {
	Enabled              bool     `envconfig:"ENABLED" default:"true" json:"enabled"`
	Port                 int      `envconfig:"PORT" default:"20000" json:"port"`
	CORS                 bool     `envconfig:"CORS" default:"false" json:"cors"`
	CORSAllowedOrigins   []string `envconfig:"CORS_ALLOWED_ORIGINS" json:"cors_allowed_origins"`
	CORSAllowCredentials bool     `envconfig:"CORS_ALLOW_CREDENTIALS" default:"false" json:"cors_allow_credentails"`
	AccessLogInfoLevel   bool     `envconfig:"ACCESS_LOG_INFO" default:"false" json:"access_log_info"`
}

type grpcCfg struct {
	Enabled bool `envconfig:"ENABLED" default:"true" json:"enabled"`
	Port    int  `envconfig:"PORT" default:"20001" json:"port"`
}

type SvcCfg struct {
	Log       Z.Config `envconfig:"LOG" json:"log"`
	HTTP      httpCfg  `envconfig:"HTTP" json:"http"`
	GRPC      grpcCfg  `envconfig:"GRPC" json:"grpc"`
	PprofPort int      `envconfig:"PPROF_PORT" json:"pprof_port"`
}

func loadCfg(l L.L, name string, dst interface{}, path string) error {
	l = l.With("name", name)

	// defaults are fetched from env. overriden by config file if path supplied.
	if err := envconfig.Process(name, dst); err != nil {
		return fmt.Errorf("config load error: %w", err)
	}

	if path != "" {
		l.Info("loading config from file", "path", path)

		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %q: %w", path, err)
		}

		if err := yaml.Unmarshal(bs, dst); err != nil {
			return fmt.Errorf("parse %q: %w", path, err)
		}
	}

	if post, ok := dst.(interface{ PostSvcLoad(L.L) error }); ok {
		if err := post.PostSvcLoad(l); err != nil {
			return fmt.Errorf("post: %w", err)
		}
	}

	return nil
}

func loadSvcCfg(name, path string) (*SvcCfg, error) {
	var cfg SvcCfg

	if err := loadCfg(&L.Nullable{}, name, &cfg, path); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func printUsage(name string, cfgs ...interface{}) {
	const format = `{{range .}}{{usage_key .}}	{{usage_type .}}	{{usage_default .}}	{{usage_required .}}	{{usage_description .}}
{{end}}`

	tabs := tabwriter.NewWriter(os.Stdout, 1, 0, 4, ' ', 0)

	fmt.Fprintln(tabs, "KEY	TYPE	DEFAULT	REQUIRED	DESCRIPTION")

	_ = envconfig.Usagef(name, &SvcCfg{}, tabs, format)

	for _, cfg := range cfgs {
		_ = envconfig.Usagef(name, cfg, tabs, format)
	}

	tabs.Flush()
}
