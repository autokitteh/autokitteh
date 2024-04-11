package systest

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/tests/internal/svcproc"
)

var useProcSvc, _ = strconv.ParseBool(os.Getenv("AK_SYSTEST_USE_PROC_SVC"))

// Start the AK server, but in a goroutine rather than as a separate
// subprocess: to support breakpoint debugging, and measure test coverage.
func startAKServer(ctx context.Context, akPath string) (svc.Service, error) {
	cfg := kittehs.Must1(svc.LoadConfig("", map[string]any{
		"db.type":   "sqlite",
		"db.dsn":    "sqlite:autokitteh.sqlite",
		"http.addr": ":0",
	}, ""))

	runOpts := svc.RunOptions{Mode: "test"}

	var (
		service svc.Service
		err     error
	)

	if useProcSvc {
		service, err = svcproc.NewSvcProc(akPath, cfg, runOpts)
	} else {
		service, err = svc.New(cfg, runOpts)
	}

	if err != nil {
		return nil, fmt.Errorf("svc.New: %w", err)
	}

	if err := service.Start(ctx); err != nil {
		panic(fmt.Errorf("fx app start: %w", err))
	}

	return service, nil
}
