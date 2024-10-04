package svc

import (
	"go.autokitteh.dev/autokitteh/internal/backend/config"
)

type Config = config.Config

const (
	Default = config.Default
	Dev     = config.Dev
	Test    = config.Test
)

var ParseMode = config.ParseMode

const ConfigDelim = config.Delim
