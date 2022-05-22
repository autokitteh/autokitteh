package svc

import (
	"github.com/autokitteh/L"
)

type callback struct {
	n string
	f interface{}
}

func callbacksNames(cbs []callback) []string {
	ns := make([]string, len(cbs))
	for i, cb := range cbs {
		ns[i] = cb.n
	}
	return ns
}

type Flags struct {
	ConfigPath                                      string
	Enables, Disables, Onlys, Excepts               []string
	Setup, HelpConfig, PrintConfig, ExitBeforeStart bool
}

type opts struct {
	name                          string
	cfgs                          []interface{}
	inits, setups, starts, readys []callback
	providers                     []interface{}
	grpc, http                    bool
	defaultDisables               []string
	flags                         *Flags
	l                             func() L.L

	cli []CLIOptFunc
}

type OptFunc func(*opts)

// Set service name, the executable name by default.
func WithName(name string) OptFunc { return func(c *opts) { c.name = name } }

// Load cfg from environment using envconfig before anything else.
func WithConfig(cfg interface{}) OptFunc { return func(c *opts) { c.cfgs = append(c.cfgs, cfg) } }

// Replace command line flags with Flags.
func WithFlags(f *Flags) OptFunc { return func(c *opts) { c.flags = f } }

func WithLogger(l func() L.L) OptFunc { return func(c *opts) { c.l = l } }

func WithCLIOptions(fs ...CLIOptFunc) OptFunc {
	return func(c *opts) { c.cli = append(c.cli, fs...) }
}

// NOTE ABOUT FUNCTIONS IN WITH*:
//
// Init functions are called first. Then Setup functions if -setup is specified.
// Then Start functions. Functions of the same type (Init, Setup or Start) are
// called in the order they were specified.
//
// Any function supplied to a With* function will be called with arguments
// that it specifies, if available. If an argument is not available, its zero
// value is set.
//
// Available arguments are fulfilled from a list of "providers". A provider is
// a value that was either returned from either of the specified functions or
// explicitly provided using the Provide function below.
//
// Providers that are always available:
// - context.Context
// - *zap.SugaredLogger
// - Configuration as was supplied to WithConfig.
//
// For example:
//
//   WithInit("example", func(z *zap.SugaredLogger) SomeType {
//     z.Info("hi there!")
//     return SomeType{...}
//   }),
//   WithStart("example", func(z *zap.SugaredLogger, st SomeType) {
//     z.Info("oh hello again!")
//     ...
//   })
//
// Note that arguments are fulfilled by type, so don't use common types such as
// int, string, etc to select them. Instead wrap them, for example:
//
//   type StartupTime time.Time
//
// Errors returned from functions are never used as providers.

// Call the init function f after log is initialized and config is loaded.
func WithInit(n string, f interface{}) OptFunc {
	if f == nil {
		return func(*opts) {}
	}

	return func(c *opts) { c.inits = append(c.inits, callback{n: n, f: f}) }
}

// Call the setup function f after init but before start. This is done only
// if -setup is specified.
func WithSetup(n string, f interface{}) OptFunc {
	if f == nil {
		return func(*opts) {}
	}

	return func(c *opts) { c.setups = append(c.setups, callback{n: n, f: f}) }
}

// Call the start function f to start things up after init and setup.
// Additional providers available for start functions: *mux.Router, *grpc.Server.
func WithStart(n string, f interface{}) OptFunc {
	if f == nil {
		return func(*opts) {}
	}

	return func(c *opts) { c.starts = append(c.starts, callback{n: n, f: f}) }
}

// Call the ready function f after everything else.
func WithReady(n string, f interface{}) OptFunc {
	if f == nil {
		return func(*opts) {}
	}

	return func(c *opts) { c.readys = append(c.readys, callback{n: n, f: f}) }
}

// A component is a grouping of init/setup/start record that conceptually
// belong to a specific component with a name.
type Component struct {
	Name                      string
	Init, Setup, Start, Ready interface{}
	Disabled                  bool
}

func WithComponent(comps ...Component) OptFunc {
	return func(c *opts) {
		for _, comp := range comps {
			WithInit(comp.Name, comp.Init)(c)
			WithSetup(comp.Name, comp.Setup)(c)
			WithStart(comp.Name, comp.Start)(c)
			WithReady(comp.Name, comp.Ready)(c)

			if comp.Disabled {
				WithDefaultDisable(comp.Name)(c)
			}
		}
	}
}

// Disable components by default unless enabled from cmdline.
func WithDefaultDisable(ns ...string) OptFunc {
	return func(c *opts) { c.defaultDisables = append(c.defaultDisables, ns...) }
}

// Explicitly provide a new provider before initialization.
func Provide(p interface{}) OptFunc {
	return func(c *opts) { c.providers = append(c.providers, p) }
}

func WithGRPC(enabled bool) OptFunc { return func(c *opts) { c.grpc = enabled } }

func WithHTTP(enabled bool) OptFunc { return func(c *opts) { c.http = enabled } }
