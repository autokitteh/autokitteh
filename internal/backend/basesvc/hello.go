package basesvc

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/fatih/color"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/version"
)

//go:embed hello.txt
var hello string

var helloTemplate = template.Must(template.New("hello").Parse(hello))

func sayHello(opts RunOptions, addr string) {
	fieldColor := color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	eyeColor := color.New(color.FgGreen).Add(color.Bold).SprintFunc()

	var mode string
	if opts.Mode != "" {
		mode = "Mode:      " + fieldColor(opts.Mode) + " "
	}

	kittehs.Must0(helloTemplate.Execute(os.Stderr, struct {
		Version string
		PID     string
		Addr    string
		Eye     string
		Mode    string
	}{
		Version: fieldColor(version.Version),
		PID:     fieldColor(fmt.Sprintf("%d", os.Getpid())),
		Addr:    fieldColor(addr),
		Eye:     eyeColor("▀"),
		Mode:    mode,
	}))
}
