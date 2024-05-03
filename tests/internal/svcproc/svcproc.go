package svcproc

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

const (
	httpAddrFile = "ak_server_addr.txt"
)

type svcProc struct {
	binPath string
	cfg     *svc.Config
	ropts   svc.RunOptions
	cmd     *exec.Cmd
	wait    chan svc.ShutdownSignal
}

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

func (s *svcProc) Start(ctx context.Context) error {
	// TODO: Pass user configuration to executable.
	s.cmd = exec.Command(
		s.binPath, "up",
		"--mode", string(s.ropts.Mode),

		"--config", "db.type=sqlite",
		"--config", "db.dsn=file:autokitteh.sqlite", // In the test's temporary directory.

		"--config", "http.addr=:0",
		"--config", "http.addr_filename="+httpAddrFile, // In the test's temporary directory.
	)
	// Use same system group to kill ak + all children (temporal, etc.)
	s.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("start subprocess: %w", err)
	}

	s.wait = make(chan fx.ShutdownSignal, 1)
	go func() {
		var exitCode int

		if err := s.cmd.Wait(); err != nil {
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = 1010 // Indicates some kind of IO error.
			}
		}

		// TODO: Pass which signal killed it somehow?
		s.wait <- svc.ShutdownSignal{ExitCode: exitCode}
	}()

	return nil
}

func (s *svcProc) Stop(ctx context.Context) error {
	// TODO: Wait until stopping is verified?

	pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
	if err == nil {
		err = syscall.Kill(-pgid, syscall.SIGTERM)
	}
	if err != nil { // If anything was wrong just kill the AK server.
		_ = s.cmd.Process.Kill()
	}
	return err
}

func (s *svcProc) Wait() <-chan svc.ShutdownSignal { return s.wait }
