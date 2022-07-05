//go:build !linux

package internalplugins

import (
	"syscall"
)

func osSpecificSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		// set process group ID
		Setpgid: true,
	}
}
