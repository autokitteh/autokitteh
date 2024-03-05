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

func New(cfg *Config, z *zap.Logger) (Client, error) {
	var tlsConfig *tls.Config
	if cfg.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(cfg.TLS.CertFilePath, cfg.TLS.KeyFilePath)
		if err != nil {
			return nil, fmt.Errorf("load x509 key pair: %w", err)
		}

		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	opts := client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
		Logger:    logur.LoggerToKV(zapadapter.New(z.WithOptions(zap.IncreaseLevel(cfg.LogLevel)))),
		ConnectionOptions: client.ConnectionOptions{
			TLS: tlsConfig,
		},
	}

	impl := &impl{z: z, cfg: cfg, done: make(chan struct{})}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	startDevServer := func() error {
		dscfg := cfg.DevServer
		dscfg.ClientOptions = &opts

		var err error
		if impl.srv, err = testsuite.StartDevServer(ctx, dscfg); err != nil {
			return fmt.Errorf("start dev server: %w", err)
		}

		z.Info("started temporal dev server", zap.String("address", impl.srv.FrontendHostPort()))

		impl.client = impl.srv.Client()

		return nil
	}

	if cfg.AlwaysStartDevServer {
		if err := startDevServer(); err != nil {
			return nil, err
		}
	} else {
		// TODO: tls and all this fun.
		var err error

		if impl.client, err = client.NewLazyClient(opts); err != nil {
			return nil, fmt.Errorf("new temporal client: %w", err)
		}

		if cfg.StartDevServerIfNotUp {
			if err := impl.healthcheck(ctx); err != nil {
				var unavailable *serviceerror.Unavailable
				if !errors.As(err, &unavailable) {
					return nil, fmt.Errorf("temporal client: %w", err)
				}

				z.Info("cannot connect to server, starting dev server")

				if err := startDevServer(); err != nil {
					return nil, err
				}
			}
		}
	}

	return impl, nil
}

func (c *impl) Temporal() client.Client { return c.client }

func (c *impl) Stop(context.Context) error {
	close(c.done)
	if c.srv != nil {
		if err := c.srv.Stop(); err != nil {
			return fmt.Errorf("stop dev server: %w", err)
		}
	}

	return nil
}

func (c *impl) healthcheck(ctx context.Context) error {
	c.z.Debug("checking health")

	if c.cfg.CheckHealthTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	_, err := c.client.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		return err
	}
	c.z.Debug("temporal reports healthy")

	return nil
}

func (c *impl) Start(context.Context) error {
	if c.cfg.CheckHealthInterval == 0 {
		c.z.Warn("periodical check health is disabled")
		return nil
	}

	var ok bool

	go func() {
		for {
			if err := c.healthcheck(context.Background()); err != nil {
				// TODO: stats.

				ok = false
				c.z.Error("temporal check health error", zap.Error(err))
			}

			if !ok {
				ok = true
				c.z.Info("temporal reports healthy")
			}

			select {
			case <-time.After(c.cfg.CheckHealthInterval):
				// nop
			case <-c.done:
				return
			}
		}
	}()

	return nil
}
