package svc

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/svc"
)

const (
	Default = configset.Default
	Dev     = configset.Dev
	Test    = configset.Test
)

var ParseMode = configset.ParseMode

const ConfigDelim = svc.Delim
