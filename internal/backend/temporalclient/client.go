package temporalclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.uber.org/zap"
	zapadapter "logur.dev/adapter/zap"
	"logur.dev/logur"

	"go.autokitteh.dev/autokitteh/internal/backend/health/healthreporter"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

type Client interface {
	Start(context.Context) error
	Stop(context.Context) error
	Temporal() client.Client
	TemporalAddr() (frontend, ui string)
	healthreporter.HealthReporter
}

type impl struct {
	client  client.Client
	z       *zap.Logger
	cfg     *Config
	srv     *testsuite.DevServer
	logFile *os.File
	done    chan struct{}
	opts    client.Options
}

func NewFromClient(cfg *MonitorConfig, z *zap.Logger, tclient client.Client) (Client, error) {
	if cfg == nil {
		cfg = &MonitorConfig{}
	}
	return &impl{z: z, cfg: &Config{Monitor: *cfg}, client: tclient, done: make(chan struct{})}, nil
}

func New(cfg *Config, z *zap.Logger) (Client, error) {
	var tlsConfig *tls.Config
	if cfg.TLS.Enabled {
		var cert tls.Certificate
		var err error
		if cfg.TLS.Certificate != "" && cfg.TLS.Key != "" {
			cert, err = tls.X509KeyPair([]byte(cfg.TLS.Certificate), []byte(cfg.TLS.Key))
		} else if cfg.TLS.CertFilePath != "" && cfg.TLS.KeyFilePath != "" {
			cert, err = tls.LoadX509KeyPair(cfg.TLS.CertFilePath, cfg.TLS.KeyFilePath)
		} else {
			return nil, errors.New("tls enabled without certificate or key")
		}

		if err != nil {
			return nil, fmt.Errorf("load x509 key pair: %w", err)
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	return &impl{
		z:    z,
		cfg:  cfg,
		done: make(chan struct{}),
		opts: client.Options{
			HostPort:  cfg.HostPort,
			Namespace: cfg.Namespace,
			Logger:    logur.LoggerToKV(zapadapter.New(z.WithOptions(zap.IncreaseLevel(cfg.Monitor.LogLevel)))),
			ConnectionOptions: client.ConnectionOptions{
				TLS: tlsConfig,
			},
		},
	}, nil
}

func (c *impl) startDevServer(ctx context.Context) error {
	var err error
	logPath := path.Join(xdg.DataHomeDir(), "temporal_dev.log")
	c.logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open Temporal dev server log file: %w", err)
	}

	c.cfg.DevServer.ClientOptions = &c.opts
	c.cfg.DevServer.Stderr = c.logFile
	c.cfg.DevServer.Stdout = c.logFile

	if c.srv, err = testsuite.StartDevServer(ctx, c.cfg.DevServer); err != nil {
		return fmt.Errorf("start Temporal dev server: %w", err)
	}

	c.z.Info("Started Temporal dev server", zap.String("address", c.srv.FrontendHostPort()))

	if c.client != nil {
		c.client.Close()
	}

	c.client = c.srv.Client()

	return nil
}

func (c *impl) Temporal() client.Client { return c.client }

func (c *impl) TemporalAddr() (frontend, ui string) {
	if c.srv == nil {
		frontend = c.cfg.HostPort

		if frontend == "" {
			// known temporal defaults.
			frontend = "localhost:7233"
			ui = "http://localhost:8233"
		}

		return
	}

	frontend = c.srv.FrontendHostPort()

	if c.cfg.DevServer.EnableUI {
		host, port, err := net.SplitHostPort(frontend)
		if err != nil {
			return
		}

		nport, err := strconv.Atoi(port)
		if err != nil {
			return
		}

		// temporal's default is frontend+1000 for dev server.
		nport += 1000

		ui = fmt.Sprintf("http://%s:%d", host, nport)
	}

	return
}

func (c *impl) Stop(context.Context) error {
	close(c.done)

	if c.client != nil {
		defer c.logFile.Close()
		c.client.Close()
	}

	if c.srv != nil {
		if err := c.srv.Stop(); err != nil {
			// This is an ugly but reasonable hack: we can't do anything
			// at this point if the Temporal server's pipe is broken.
			if err.Error() == "signal: broken pipe" {
				return nil
			}
			return fmt.Errorf("stop Temporal dev server: %w", err)
		}
	}

	return nil
}

func (c *impl) healthCheck(ctx context.Context) error {
	if c.cfg.Monitor.CheckHealthTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	_, err := c.client.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		return err
	}

	return nil
}

func (c *impl) Start(context.Context) error {
	if c.client != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	if c.cfg.AlwaysStartDevServer {
		if err := c.startDevServer(ctx); err != nil {
			return err
		}
	} else {
		var err error

		if c.client, err = client.NewLazyClient(c.opts); err != nil {
			return fmt.Errorf("new temporal client: %w", err)
		}

		if c.cfg.StartDevServerIfNotUp {
			if err := c.healthCheck(ctx); err != nil {
				var unavailable *serviceerror.Unavailable
				if !errors.As(err, &unavailable) {
					return fmt.Errorf("temporal client: %w", err)
				}

				c.z.Info("Cannot connect to Temporal, starting Temporal dev server")

				if err := c.startDevServer(ctx); err != nil {
					return err
				}
			}
		}
	}

	// Start health check monitor.

	if c.cfg.Monitor.CheckHealthInterval == 0 {
		c.z.Warn("Periodic Temporal health checks are disabled")
		return nil
	}

	go func() {
		ok := false
		for {
			err := c.healthCheck(context.Background())
			if err == nil && !ok {
				c.z.Info("Connection to Temporal is healthy")
				ok = true
			} else if err != nil {
				// TODO: stats.
				c.z.Error("Temporal health check error", zap.Error(err))
				ok = false
			}

			select {
			case <-time.After(c.cfg.Monitor.CheckHealthInterval):
				// nop
			case <-c.done:
				return
			}
		}
	}()

	return nil
}

func (c *impl) Report() error {
	return c.healthCheck(context.Background())
}
