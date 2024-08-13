package remotert

import "errors"

type RemoteRuntimeConfig struct {
	RunnerAddress []string
}

func (c RemoteRuntimeConfig) validate() error {
	if len(c.RunnerAddress) == 0 {
		return errors.New("no runner manager address")
	}

	return nil
}
