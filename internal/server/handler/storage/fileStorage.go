package storage

import (
	"encoding/json"
	"os"
	"time"

	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	"github.com/megaded/metrictmr/internal/server/handler/config"
)

type FileStorage struct {
	m        Storager
	filePath string
	internal int
	restore  bool
}

func (s *FileStorage) GetGauge(name string) (metric data.Metric, exist bool) {
	return s.m.GetGauge(name)
}
func (s *FileStorage) Store(metric ...data.Metric) {
	s.m.Store(metric...)
	if s.internal == 0 {
		s.persistData()
	}
}

func (s *FileStorage) GetCounter(name string) (metric data.Metric, exist bool) {
	return s.m.GetCounter(name)
}

func (s *FileStorage) GetGaugeMetrics() []data.Metric {
	return s.m.GetGaugeMetrics()
}
func (s *FileStorage) GetCounterMetrics() []data.Metric {
	return s.m.GetCounterMetrics()
}
func (s *FileStorage) HealthCheck() bool {
	return true
}

func NewFileStorage(cfg config.Config) *FileStorage {
	fs := FileStorage{m: NewInMemoryStorage(), internal: *cfg.StoreInterval, filePath: cfg.FilePath, restore: *cfg.Restore}
	if fs.internal != 0 {
		go func() {

			count := 0
			for {
				if count%*cfg.StoreInterval == 0 {
					fs.persistData()
				}
				time.Sleep(time.Second)
				count++
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

func (s *FileStorage) persistData() {
	gaugeMetrics := s.m.GetGaugeMetrics()
	counterMetrics := s.m.GetCounterMetrics()
	storeData := make([]data.Metric, 0)
	if len(gaugeMetrics) != 0 {
		storeData = append(storeData, gaugeMetrics...)
	}
	if len(counterMetrics) != 0 {
		storeData = append(storeData, counterMetrics...)
	}
	if len(storeData) == 0 {
		return
	}
	data, err := json.Marshal(storeData)
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}
	file, err := os.Create(s.filePath)
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}
	defer file.Close()
	file.Write(data)
}
