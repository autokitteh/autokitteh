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

		if f != nil {
			f(c)
		}
		c.Logger = l
		// TODO(ENG-1590): maybe we could clone the core with the same encoder and writer, but with different/lower level?
		// (we cannon lower level with zap.IncreaseLevel(), unfortunately or extract encoder and writer from the original core)
		// so, if explicit debug is requested - switch to default gorm logger (which will log to stdout)
		if cfg.Debug {
			c.Logger = logger.Default
			l.LogLevel = logger.Info
		}
	})
}
