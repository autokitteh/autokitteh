package usagereporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/version"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

type reportRequest struct {
	InstallationID string `json:"installation_id"`
	Version        string `json:"version"`
	Commit         string `json:"commit"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
}

type UsageReporter interface {
	Start()
	Stop()
}

type usageReporter struct {
	installationID uuid.UUID
	updateInterval time.Duration
	shutdownChan   chan struct{}
	endpoint       string
	logger         *zap.Logger
}

func generateInstallIDFile(path string) error {
	id := uuid.New()
	if err := os.WriteFile(path, []byte(id.String()), 0o600); err != nil {
		return err
	}
	return nil
}

func ensureAndReadInstallationIDFile(installationIDFile string) (string, error) {
	// Try read installationID file,
	// if there is an error and the file exists, try to remove it and fail if cant
	// generate a new file
	// try read the data again
	data, err := os.ReadFile(installationIDFile)
	if err != nil {
		if !os.IsNotExist(err) {
			if err := os.Remove(installationIDFile); err != nil {
				return "", err
			}
		}

		if err := generateInstallIDFile(installationIDFile); err != nil {
			return "", err
		}
		if data, err = os.ReadFile(installationIDFile); err != nil {
			return "", err
		}

	}

	return string(data), nil
}

func New(z *zap.Logger, cfg *Config) (UsageReporter, error) {
	if !cfg.Enabled {
		return &nopUpdater{}, nil
	}

	installationIDFile := filepath.Join(xdg.DataHomeDir(), "installID")
	data, err := ensureAndReadInstallationIDFile(installationIDFile)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(data)
	if err != nil {
		data, err = ensureAndReadInstallationIDFile(installationIDFile)
		if err != nil {
			return nil, err
		}
		id, err = uuid.Parse(data)
		if err != nil {
			return nil, err
		}
	}

	return &usageReporter{
		installationID: id,
		updateInterval: cfg.Interval,
		shutdownChan:   make(chan struct{}),
		endpoint:       cfg.Endpoint,
		logger:         z,
	}, nil
}

func (d *usageReporter) report() {
	r := reportRequest{
		InstallationID: d.installationID.String(),
		Version:        version.Version,
		Commit:         version.Commit,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
	}

	data, err := json.Marshal(r)
	if err != nil {
		d.logger.Debug("report usage data failed", zap.Error(err))
		return
	}

	if err := post(d.endpoint, data); err != nil {
		d.logger.Debug("report usage data failed", zap.Error(err))
		return
	}
	d.logger.Debug("report usage data succeed")
}

func (d *usageReporter) Start() {
	go func() {
		d.logger.Debug("start usage updating loop")
		d.report()
		for {
			select {
			case <-d.shutdownChan:
				d.logger.Debug("stopped usage updating loop")
				return
			case <-time.After(d.updateInterval):
				d.report()
			}
		}
	}()
}

func (d *usageReporter) Stop() {
	close(d.shutdownChan)
}
