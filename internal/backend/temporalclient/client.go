package temporalclient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.uber.org/zap"
	zapadapter "logur.dev/adapter/zap"
	"logur.dev/logur"
)

type Client interface {
	Start(context.Context) error
	Stop(context.Context) error
	Temporal() client.Client
}

type impl struct {
	client client.Client
	z      *zap.Logger
	cfg    *Config
	srv    *testsuite.DevServer
	done   chan struct{}
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
			cert, err = tls.LoadX509KeyPair(cfg.TLS.CertFilePath, cfg.TLS.KeyFilePath)
		} else if cfg.TLS.CertFilePath != "" && cfg.TLS.KeyFilePath != "" {
			cert, err = tls.X509KeyPair([]byte(cfg.TLS.Certificate), []byte(cfg.TLS.Key))
		} else {
			return nil, errors.New("tls enabled without certificate or key")
		}

		if err != nil {
			return nil, fmt.Errorf("load x509 key pair: %w", err)
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	opts := client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
		Logger:    logur.LoggerToKV(zapadapter.New(z.WithOptions(zap.IncreaseLevel(cfg.Monitor.LogLevel)))),
		ConnectionOptions: client.ConnectionOptions{
			TLS: tlsConfig,
		},
	}

	impl := &impl{z: z, cfg: cfg, done: make(chan struct{})}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	if cfg.AlwaysStartDevServer {
		if err := impl.startDevServer(ctx, cfg, opts); err != nil {
			return nil, err
		}
	} else {
		// TODO: tls and all this fun.
		var err error

		if impl.client, err = client.NewLazyClient(opts); err != nil {
			return nil, fmt.Errorf("new temporal client: %w", err)
		}

		if cfg.StartDevServerIfNotUp {
			if err := impl.healthCheck(ctx); err != nil {
				var unavailable *serviceerror.Unavailable
				if !errors.As(err, &unavailable) {
					return nil, fmt.Errorf("temporal client: %w", err)
				}

				z.Info("Cannot connect to Temporal, starting Temporal dev server")

				if err := impl.startDevServer(ctx, cfg, opts); err != nil {
					return nil, err
				}
			}
		}
	}

	return impl, nil
}

func (c *impl) startDevServer(ctx context.Context, cfg *Config, opts client.Options) error {
	cfg.DevServer.ClientOptions = &opts

	var err error
	if c.srv, err = testsuite.StartDevServer(ctx, cfg.DevServer); err != nil {
		return fmt.Errorf("start Temporal dev server: %w", err)
	}
	c.z.Info("Started Temporal dev server", zap.String("address", c.srv.FrontendHostPort()))

	c.client = c.srv.Client()

	return nil
}

func (c *impl) Temporal() client.Client { return c.client }

func (c *impl) Stop(context.Context) error {
	close(c.done)

	if c.client != nil {
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
	c.z.Debug("Checking Temporal connection health")

	if c.cfg.Monitor.CheckHealthTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	_, err := c.client.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		return err
	}
	c.z.Debug("Connection to Temporal is healthy")

	return nil
}

func (c *impl) Start(context.Context) error {
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
