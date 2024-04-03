package usageupdater

type nopUpdater struct{}

func (nopUpdater) Start() {}
func (nopUpdater) Stop()  {}
