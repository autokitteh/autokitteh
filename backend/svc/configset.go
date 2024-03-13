package svc

import (
	"go.autokitteh.dev/autokitteh/internal/backend/basesvc"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

const (
	Default = configset.Default
	Dev     = configset.Dev
	Test    = configset.Test
)

var ParseMode = configset.ParseMode

const ConfigDelim = basesvc.Delim
