package systest

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

// Start the AK server, but in a goroutine rather than as a separate
// subprocess: to support breakpoint debugging, and measure test coverage.
func startAKServer(ctx context.Context) (svc.Service, error) {
	cfg := kittehs.Must1(svc.LoadConfig("", map[string]any{
		"db.dsn":    "sqlite:test.sqlite",
		"http.addr": ":0",
	}, ""))

	svc, err := svc.New(cfg, svc.RunOptions{Mode: "test"})
	if err != nil {
		return nil, fmt.Errorf("svc.New: %w", err)
	}

	if err := svc.Start(ctx); err != nil {
		panic(fmt.Errorf("fx app start: %w", err))
	}

	return svc, nil
}
