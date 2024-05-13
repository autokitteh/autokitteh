package healthreporter

type HealthReporter interface {
	Report() error
}
