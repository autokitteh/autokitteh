//go:build !windows

package temporaldevsrv

import (
	"os"
	"os/exec"
	"syscall"
)

// newCmd creates a new command with the given executable path and arguments.
func newCmd(exePath string, args ...string) *exec.Cmd {
	cmd := exec.Command(exePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// sendInterrupt sends an interrupt signal to the given process for graceful shutdown.
func sendInterrupt(process *os.Process) error {
	return process.Signal(syscall.SIGINT)
}
