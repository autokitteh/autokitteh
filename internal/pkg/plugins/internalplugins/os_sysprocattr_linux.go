//go:build linux

package internalplugins

import (
	"syscall"
)

func osSpecificSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		// kill children if parent is dead
		Pdeathsig: syscall.SIGKILL,
		// set process group ID
		Setpgid: true,
	}
}
