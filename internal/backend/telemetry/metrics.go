package telemetry

var Metrics metrics

type metrics struct {
	Sessions UpDownCounter
}
