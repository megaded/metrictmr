package storage

import (
	"context"

	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/server/handler/config"
)

type Storager interface {
	GetGauge(name string) (metric data.Metric, exist bool, err error)
	Store(metric ...data.Metric) error
	GetCounter(name string) (metric data.Metric, exist bool, err error)
	GetMetrics() ([]data.Metric, error)
	HealthCheck() bool
}

func CreateStorage(ctx context.Context, cfg config.Config) Storager {
	if cfg.DBConnString != "" {
		return NewPgStorage(ctx, cfg)
	}
	_, isDefault := cfg.GetFilePath()
	if !isDefault {
		return NewFileStorage(ctx, cfg)
	}
	return NewInMemoryStorage()
}
