//go:build temporalite

package aksvc

// [# with_temporalite #]

/*
Requires this in go.mod:

require	github.com/DataDog/temporalite v0.0.0-20220126212208-de413afe117f

replace (
	github.com/apache/thrift => github.com/apache/thrift v0.0.0-20190309152529-a9b748bb0e02
	github.com/cactus/go-statsd-client => github.com/cactus/go-statsd-client v3.2.1+incompatible
	go.temporal.io/server => go.temporal.io/server v1.13.1-0.20211229212730-1e02664377f6
)

import (
	"github.com/autokitteh/autokitteh/internal/app/temporalite"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/pkg/svc"
)

var TemporaliteComponent = svc.Component{
	Name: "temporalite",
	Init: func(l L.L, cfg *Config) error {
		return temporalite.Start(l, cfg.Temporalite)
	},
}

*/
