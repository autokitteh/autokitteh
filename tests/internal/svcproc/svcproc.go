package svcproc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type svcProc struct {
	binPath string
	ropts   svc.RunOptions
	cfg     *svc.Config
	addr    string
	cmd     *exec.Cmd
	wait    chan svc.ShutdownSignal
}

func (s *svcProc) Start(ctx context.Context) error {
	tmpDir, err := os.MkdirTemp("", "ak-svcproc-*")
	if err != nil {
		return fmt.Errorf("mkdirtemp: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "rm(%q): %v\n", tmpDir, err)
		}
	}()

	addrFilePath := filepath.Join(tmpDir, "addr")
	readyFilePath := filepath.Join(tmpDir, "ready")

	// TODO: pass user configuration to executable.
	cmd := exec.Command(
		s.binPath,
		"--config", "http.addr=:0",
		"--config", fmt.Sprintf("http.addr_filename=%s", addrFilePath),
		"up",
		"-m", string(s.ropts.Mode),
		"-r", readyFilePath,
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	wait := make(chan fx.ShutdownSignal, 1)
	go func() {
		var rc int

		if err := cmd.Wait(); err != nil {
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				rc = exitError.ExitCode()
			} else {
				rc = 1010 // indicates some kind of IO error.
			}
		}

		// TODO: pass which signal killed it somehow?
		wait <- svc.ShutdownSignal{ExitCode: rc}
	}()

	timeout := time.After(5 * time.Second)

	for ready := false; !ready; {
		select {
		case <-wait:
			return errors.New("service stopped before being ready")
		case <-time.After(100 * time.Millisecond):
			if _, err := os.Stat(addrFilePath); err == nil {
				ready = true
			}
		case <-timeout:
			return errors.New("timeout waiting for service to start")
		}
	}

	addr, err := os.ReadFile(addrFilePath)
	if err != nil {
		return fmt.Errorf("read addr: %w", err)
	}

	s.addr = strings.TrimSpace(string(addr))
	s.wait = wait
	s.cmd = cmd

	return nil
}

func (s *svcProc) Stop(ctx context.Context) error {
	// TODO: wait until verified stopped?
	return s.cmd.Process.Kill()
}

func (s *svcProc) Wait() <-chan svc.ShutdownSignal { return s.wait }

func (s *svcProc) Addr() string { return s.addr }

func NewSvcProc(binPath string, cfg *svc.Config, ropts svc.RunOptions) (svc.Service, error) {
	if ropts.TemporalClient != nil {
		return nil, sdkerrors.ErrNotImplemented
	}

	return &svcProc{
		binPath: binPath,
		cfg:     cfg,
		ropts:   ropts,
	}, nil
}
