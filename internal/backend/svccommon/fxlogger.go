package svccommon

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

func LoggerFxOpt(silent bool) fx.Option {
	if silent {
		return fx.Supply(zap.NewNop())
	}

	return fx.Module(
		"logger",
		fx.Provide(fxGetConfig("logger", kittehs.Must1(chooseConfig(logger.Configs)))),
		fx.Provide(logger.New),
	)
}
