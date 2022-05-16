package z

import "go.uber.org/zap"

var znop = zap.NewNop().Sugar()

// If z is nil, nop logger is returned. Otherwise z is returned.
func Z(z *zap.SugaredLogger) *zap.SugaredLogger {
	if z == nil {
		return znop
	}

	return z
}
