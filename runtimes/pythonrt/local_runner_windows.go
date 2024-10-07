//go:build windows

package pythonrt

import "os/exec"

func setCmdSysAttrPGID(cmd *exec.Cmd) {}
