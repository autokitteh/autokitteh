package gormkitteh

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

func OpenZ(z *zap.Logger, cfg *Config, f func(*gorm.Config)) (*gorm.DB, error) {
	if z == nil {
		return Open(cfg, f)
	}

	return Open(cfg, func(c *gorm.Config) {
		l := &zapgorm2.Logger{
			ZapLogger:                 z,
			LogLevel:                  logger.Warn,
			SlowThreshold:             cfg.SlowQueryThreshold,
			IgnoreRecordNotFoundError: true,
		}

		if cfg.Debug {
			l.LogLevel = logger.Info
		}

		c.Logger = l

		if f != nil {
			f(c)
		}
	})
}
