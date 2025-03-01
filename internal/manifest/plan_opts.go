package manifest

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type opts struct {
	fromScratch      bool
	log              Log
	projectName      string
	oid              sdktypes.OrgID
	rmUnusedConnVars bool
}

func applyOptions(optfns []Option) (opts opts) {
	for _, fn := range optfns {
		fn(&opts)
	}
	return
}

type Option func(*opts)

func WithFromScratch(s bool) Option           { return func(o *opts) { o.fromScratch = s } }
func WithRemoveUnusedConnFlags(s bool) Option { return func(o *opts) { o.rmUnusedConnVars = s } }
func WithLogger(l Log) Option                 { return func(o *opts) { o.log = l } }
func WithProjectName(n string) Option         { return func(o *opts) { o.projectName = n } }
func WithOrgID(id sdktypes.OrgID) Option      { return func(o *opts) { o.oid = id } }
