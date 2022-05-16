package clitools

import "fmt"

var addr string

func Addr() string { return addr }

func Init(addr_ string) error {
	addr = addr_

	if err := initLog(Settings.LogLevel); err != nil {
		return fmt.Errorf("log: %w", err)
	}

	return nil
}
