//go:build !windows

package pythonrt

import (
	"os/exec"
	"syscall"
)

func setCmdSysAttrPGID(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
