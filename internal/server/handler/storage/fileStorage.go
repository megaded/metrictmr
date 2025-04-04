package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/retry"
	"github.com/megaded/metrictmr/internal/server/handler/config"
)

type FileStorage struct {
	m        Storager
	filePath string
	internal int
	restore  bool
	retry    retry.Retry
}

func (s *FileStorage) GetGauge(name string) (metric data.Metric, exist bool, err error) {
	return s.m.GetGauge(name)
}
func (s *FileStorage) Store(metric ...data.Metric) error {
	err := s.m.Store(metric...)
	if err != nil {
		return err
	}
	if s.internal == 0 {
		return s.persistData()
	}
	return nil
}

func (s *FileStorage) GetCounter(name string) (metric data.Metric, exist bool, err error) {
	return s.m.GetCounter(name)
}

func (s *FileStorage) GetMetrics() ([]data.Metric, error) {
	return s.m.GetMetrics()
}
func (s *FileStorage) HealthCheck() bool {
	return true
}

func NewFileStorage(ctx context.Context, cfg config.Config) *FileStorage {
	fs := FileStorage{m: NewInMemoryStorage(), internal: *cfg.StoreInterval, filePath: cfg.FilePath, restore: *cfg.Restore, retry: retry.NewRetry(1, 2, 3)}
	if fs.internal != 0 {
		go func() {
			timer := time.NewTicker(time.Duration(*cfg.StoreInterval * int(time.Second)))
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					fs.persistData()
				}
			}
		}()
	}
	if fs.restore {
		fs.restoreStorage()
	}
	return &fs
}

func (s FileStorage) restoreStorage() {
	var metrics []data.Metric
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}

	err = json.Unmarshal(data, &metrics)
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}
	for _, m := range metrics {
		s.Store(m)
	}
}

func (s *FileStorage) persistData() error {
	metrics, err := s.m.GetMetrics()
	if err != nil {
		return err
	}

	if len(metrics) == 0 {
		return nil
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Info(err.Error())
		return err
	}
	file, err := os.Create(s.filePath)
	if err != nil {
		logger.Log.Info(err.Error())
		return err
	}
	defer file.Close()
	action := func() error {
		d, err := file.Write(data)
		if err != nil {
			return err
		}
		if d != len(data) {
			return errors.New("invalid write data")
		}
		defer file.Close()
		return nil
	}
	fn := s.retry.Retry(context.TODO(), action)
	return fn()
}
