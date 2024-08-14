package remotert

import "errors"

type RemoteRuntimeConfig struct {
	ManagerAddress []string
}

func (c RemoteRuntimeConfig) validate() error {
	if len(c.ManagerAddress) == 0 {
		return errors.New("no runner manager address")
	}

	return nil
}
