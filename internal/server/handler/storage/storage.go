package storage

import (
	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/server/handler/config"
)

type Storager interface {
	GetGauge(name string) (metric data.Metric, exist bool)
	Store(metric ...data.Metric)
	GetCounter(name string) (metric data.Metric, exist bool)
	GetGaugeMetrics() []data.Metric
	GetCounterMetrics() []data.Metric
	HealthCheck() bool
}

func CreateStorage(cfg config.Config) Storager {
	if cfg.DBConnString != "" {
		return NewPgStorage(cfg)
	}
	_, isDefault := cfg.GetFilePath()
	if !isDefault {
		return NewFileStorage(cfg)
	}
	return NewInMemoryStorage()
}
