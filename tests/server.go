package tests

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

type AKServer struct {
	cmd  *exec.Cmd
	log  *os.File
	Addr string
}

const (
	addrFilename = "ak_addr"
	logFilename  = "ak_server.log"
	startTimeout = 10 * time.Second
)

// StartAKServer starts the AK server as a subprocess.
// If the server started but isn't responding, we still return
// the process details, so the caller can kill the entire process group.
// It is assumed that the AK binary was built before running the test,
// and that the current working directory is temporary and isolated.
func StartAKServer(akPath, akMode string) (*AKServer, error) {
	cmd := exec.Command(akPath, "up", "--mode", akMode)

	// Associate AK's child processes (e.g. Temporal) with the
	// same process group as AK, to kill them all together.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Server configuration.
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "AK_DB__TYPE=sqlite")
	cmd.Env = append(cmd.Env, "AK_DB__DSN=file:autokitteh.sqlite")
	cmd.Env = append(cmd.Env, "AK_HTTP__ADDR=:0")
	cmd.Env = append(cmd.Env, "AK_HTTP__ADDR_FILENAME="+addrFilename)
	cmd.Env = append(cmd.Env, "AK_LOGGER__LEVEL=info")
	cmd.Env = append(cmd.Env, "AK_PPROF__ENABLE=false")
	cmd.Env = append(cmd.Env, "AK_PYTHONRT__LAZY_LOAD_LOCAL_VENV=false")
	cmd.Env = append(cmd.Env, "AK_TEMPORALCLIENT__ALWAYS_START_DEV_SERVER=true")
	cmd.Env = append(cmd.Env, "AK_WEBPLATFORM__PORT=0")

	// Capture stdout and stderr in a log file.
	log, err := os.OpenFile(logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create AK server log file: %w", err)
	}

	cmd.Stdout = log
	cmd.Stderr = log

	// Start the AK server.
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start AK server: %w", err)
	}

	server := &AKServer{cmd: cmd, log: log}

	// Wait for it to be ready.
	addr := waitForAddress()
	if addr == "" {
		return server, errors.New("timed out waiting for AK server address")
	}

	server.Addr = addr
	return server, nil
}

// waitForAddress waits for the AK server to write its address to a file.
// It polls the file every 0.1s, and returns an empty string if it times out.
func waitForAddress() string {
	done := make(chan string)
	timeout := time.After(startTimeout)
	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if addr, err := os.ReadFile(addrFilename); err == nil {
					done <- string(addr)
					return
				}
			case <-timeout:
				done <- ""
				return
			}
		}
	}()

	return <-done
}

// PrintLog prints the AK server log from a temporary file to the test log.
// Called after test errors, to help diagnose them but keep the test log manageable.
func (s *AKServer) PrintLog(t *testing.T) {
	if s == nil {
		return
	}

	_ = s.log.Sync()
	_ = s.log.Close()

	log, err := os.ReadFile(logFilename)
	if err != nil {
		t.Log("failed to read AK server log:", err)
		return
	}

	t.Log(string(log))
}

// Stop kills the AK server and all its child processes, if possible.
func (s *AKServer) Stop() {
	if s == nil {
		return
	}

	pgid, err := syscall.Getpgid(s.cmd.Process.Pid)
	if err == nil {
		err = syscall.Kill(-pgid, syscall.SIGTERM)
	}

	// If killing the process group failed, at least kill the AK server.
	if err != nil {
		_ = s.cmd.Process.Kill()
	}
}
