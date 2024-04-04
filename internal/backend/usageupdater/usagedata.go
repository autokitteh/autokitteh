package usageupdater

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/uuid"
	"go.autokitteh.dev/autokitteh/internal/version"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	"go.uber.org/zap"
)

type updateRequest struct {
	InstallationID string `json:"installationID"`
	Version        string `json:"version"`
	Commit         string `json:"commit"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
}

type UsageUpdater interface {
	Start()
	Stop()
}

type usageUpdater struct {
	installationID uuid.UUID
	updateInterval time.Duration
	shutdownChan   chan struct{}
	poster         poster
	logger         *zap.Logger
}

func New(z *zap.Logger, cfg *Config) (UsageUpdater, error) {
	if !cfg.Enabled {
		return &nopUpdater{}, nil
	}

	installationIDFile := filepath.Join(xdg.ConfigHomeDir(), "installID")

	if _, err := os.Stat(installationIDFile); os.IsNotExist(err) {
		id := uuid.New()
		if err := os.WriteFile(installationIDFile, []byte(id.String()), 0600); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(installationIDFile)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(string(data))
	if err != nil {
		return nil, err
	}

	return &usageUpdater{
		installationID: id,
		updateInterval: cfg.Interval,
		shutdownChan:   make(chan struct{}),
		poster:         poster{endpoint: cfg.Endpoint},
		logger:         z,
	}, nil
}

func (d *usageUpdater) update() {
	r := updateRequest{
		InstallationID: d.installationID.String(),
		Version:        version.Version,
		Commit:         version.Commit,
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
	}

	data, err := json.Marshal(r)
	if err != nil {
		return
	}

	if err := d.poster.post(data); err != nil {
		d.logger.Debug("faild updated usage data", zap.Error(err))
	}

}

func (d *usageUpdater) Start() {
	go func() {
		timer := time.NewTicker(d.updateInterval)
		defer timer.Stop()

		d.logger.Debug("start usage updating loop")
		d.update()
		for {
			select {
			case <-d.shutdownChan:
				d.logger.Debug("stopped usage updating loop")
				return
			case <-timer.C:
				d.update()
			}
		}
	}()
}

func (d *usageUpdater) Stop() {
	close(d.shutdownChan)
}
