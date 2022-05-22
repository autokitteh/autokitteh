// go:build temporalite

package temporalite

// see [# with_temporalite #]

/*
import (
	"fmt"

	"github.com/DataDog/temporalite"
	uiserver "github.com/temporalio/ui-server/server"
	uiconfig "github.com/temporalio/ui-server/server/config"
	uiserveroptions "github.com/temporalio/ui-server/server/server_options"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/temporal"

	L "github.com/autokitteh/L"
)

type logger struct{ L.L }

func tagsToPairs(tags []tag.Tag) (l []interface{}) {
	for ; len(tags) != 0; tags = tags[1:] {
		t := tags[0]

		l = append(l, t.Key(), t.Value())
	}

	return
}

func (l *logger) Debug(msg string, tags ...tag.Tag) {} // l.L.Debug(msg, tagsToPairs(tags)...) }
func (l *logger) Info(msg string, tags ...tag.Tag)  {} // l.L.Info(msg, tagsToPairs(tags)...) }
func (l *logger) Warn(msg string, tags ...tag.Tag)  { l.L.Warn(msg, tagsToPairs(tags)...) }
func (l *logger) Error(msg string, tags ...tag.Tag) { l.L.Error(msg, tagsToPairs(tags)...) }
func (l *logger) Fatal(msg string, tags ...tag.Tag) { l.L.Fatal(msg, tagsToPairs(tags)...) }

func Start(l L.L, config Config) error {
	uiOpts := uiconfig.Config{
		TemporalGRPCAddress: fmt.Sprintf(":%d", config.FrontendGRPCPort),
		Port:                config.UIPort,
		EnableUI:            config.UIEnabled,
	}

	opts := []temporalite.ServerOption{
		temporalite.WithLogger(&logger{l}),
		temporalite.WithFrontendPort(config.FrontendGRPCPort),
		temporalite.WithNamespaces(config.Namespace),
		temporalite.WithSQLitePragmas(config.SQLitePragmas),
		temporalite.WithUpstreamOptions(
			temporal.InterruptOn(temporal.InterruptCh()),
		),
	}

	if config.UIEnabled {
		opts = append(opts, temporalite.WithUI(uiserver.NewServer(uiserveroptions.WithConfig(&uiOpts))))
	}

	if config.Ephemeral {
		opts = append(opts, temporalite.WithPersistenceDisabled())
	} else {
		opts = append(opts, temporalite.WithDatabaseFilePath(config.DBPath))
	}

	s, err := temporalite.NewServer(opts...)
	if err != nil {
		return fmt.Errorf("new: %w", err)
	}

	go func() {
		if err := s.Start(); err != nil {
			l.Panic("start error", "err", err)
		}

		l.Panic("temporalite server quit")
	}()

	return nil
}

*/
