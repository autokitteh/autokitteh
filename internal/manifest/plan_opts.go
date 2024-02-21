package manifest

type opts struct {
	fromScratch bool
	log         Log
}

func applyOptions(optfns []Option) (opts opts) {
	for _, fn := range optfns {
		fn(&opts)
	}
	return
}

type Option func(*opts)

func WithFromScratch(s bool) Option {
	return func(o *opts) {
		o.fromScratch = s
	}
}

func WithLogger(l Log) Option {
	return func(o *opts) {
		o.log = l
	}
}
