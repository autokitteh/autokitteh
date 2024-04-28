// The MIT License
//
// Copyright (c) 2023 Temporal Technologies Inc.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package devserver

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// newCmd creates a new command with the given executable path and arguments.
func newCmd(exePath string, args ...string) *exec.Cmd {
	cmd := exec.Command(exePath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// isolate the process and signals sent to it from the current console
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	// TODO:
	//	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd
}

// sendInterrupt calls the break event on the given process for graceful shutdown.
func sendInterrupt(process *os.Process) error {
	dll, err := windows.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	p, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	r, _, err := p.Call(uintptr(windows.CTRL_BREAK_EVENT), uintptr(process.Pid))
	if r == 0 {
		return err
	}
	return nil
}
