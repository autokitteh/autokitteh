package svc

import (
	"go.autokitteh.dev/autokitteh/internal/backend/basesvc"
	"go.autokitteh.dev/autokitteh/internal/backend/svc"
)

type (
	RunOptions = basesvc.RunOptions
	Config     = basesvc.Config
)

var (
	New        = svc.New
	LoadConfig = basesvc.LoadConfig
	StartDB    = basesvc.StartDB
)
