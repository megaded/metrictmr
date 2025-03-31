package storage

import (
	"encoding/json"
	"os"
	"time"

	"github.com/megaded/metrictmr/internal/data"
	"github.com/megaded/metrictmr/internal/logger"
	"go.uber.org/zap"
)

type FileStorage struct {
	m        InMemoryStorage
	filePath string
	internal int
	restore  bool
}

func (s *FileStorage) GetGauge(name string) (metric data.Metric, exist bool) {
	return s.m.GetGauge(name)
}
func (s *FileStorage) Store(metric data.Metric) {
	s.m.Store(metric)
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

func NewFileStorage(internal int, fp string, restore bool) *FileStorage {
	fs := FileStorage{m: NewInMemoryStorage(), internal: internal, filePath: fp, restore: restore}
	if fs.internal != 0 {
		go func() {

			count := 0
			for {
				if count%internal == 0 {
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
	logger.Log.Info("Restore metric", zap.Int("count", len(metrics)))
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
	n, err := file.Write(data)
	if err != nil {
		logger.Log.Info(err.Error())
		return
	}
	if n != len(data) {
		return
	}
}
