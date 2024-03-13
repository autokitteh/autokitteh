package svc

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type svcProc struct {
	cfg *Config
}

func (s *svcProc) Start(ctx context.Context) error {
	cmd := exec.Command("bin/ak" /*args...*/)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	return nil
}

func (s *svcProc) Stop(ctx context.Context) error { return nil }
func (s *svcProc) Wait() <-chan ShutdownSignal    { return nil }
func (s *svcProc) Addr() string                   { return "" }

func NewSvcProc(cfg *Config, ropts RunOptions) (Service, error) {
	return &svcProc{}, nil
}
