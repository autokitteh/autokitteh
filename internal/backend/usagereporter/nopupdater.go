package usagereporter

import "context"

type nopUpdater struct{}

func (nopUpdater) StartReportLoop(ctx context.Context) error { return nil }
func (nopUpdater) StopReportLoop(ctx context.Context) error  { return nil }
func (nopUpdater) Report(payload map[string]string)          {}
