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
	fs := FileStorage{m: NewInMemoryStorage(), internal: internal, filePath: fp}
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
	return &fs
}

func restoreStorage(fp string) {

}

func (s *FileStorage) persistData() {
	gaugeMetrics := s.m.GetGaugeMetrics()
	logger.Log.Info("Метрик", zap.Int("len gauge", len(gaugeMetrics)))
	counterMetrics := s.m.GetCounterMetrics()
	logger.Log.Info("Метрик", zap.Int("len counter", len(counterMetrics)))
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
	logger.Log.Info("Метрик", zap.Int("len", len(storeData)))
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
