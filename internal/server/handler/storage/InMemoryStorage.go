package storage

import (
	"github.com/megaded/metrictmr/internal/data"
)

const (
	gauge   = "gauge"
	counter = "counter"
)

type InMemoryStorage struct {
	Metrics    map[string]data.Metric
	gaugeKey   map[string]bool
	counterKey map[string]bool
}

func (s *InMemoryStorage) GetGauge(name string) (metric data.Metric, exist bool) {
	metric, exist = s.Metrics[getKey(gauge, name)]
	return metric, exist
}

func (s *InMemoryStorage) Store(metric data.Metric) {
	if metric.MType != gauge && metric.MType != counter {
		return
	}
	key := getKey(metric.MType, metric.ID)
	if metric.MType == gauge {
		s.Metrics[key] = metric
		s.gaugeKey[key] = true
	} else {
		s.storeCounter(metric)
	}

}

func (s *InMemoryStorage) GetCounter(name string) (metric data.Metric, exist bool) {
	metric, exist = s.Metrics[getKey(counter, name)]
	return metric, exist
}

func NewInMemoryStorage() InMemoryStorage {
	return InMemoryStorage{Metrics: map[string]data.Metric{}, gaugeKey: map[string]bool{}, counterKey: map[string]bool{}}
}

func (s *InMemoryStorage) GetGaugeMetrics() []data.Metric {
	result := make([]data.Metric, 0)
	for k := range s.gaugeKey {
		m, ok := s.Metrics[k]
		if ok {
			result = append(result, m)
		}
	}
	return result
}

func (s *InMemoryStorage) GetCounterMetrics() []data.Metric {
	result := make([]data.Metric, 0)
	for k := range s.counterKey {
		m, ok := s.Metrics[k]
		if ok {
			result = append(result, m)
		}

	}
	return result
}

func (s *InMemoryStorage) HealthCheck() bool {
	return true
}

func (s *InMemoryStorage) storeCounter(metric data.Metric) {
	key := getKey(counter, metric.ID)
	v, ok := s.Metrics[key]
	if ok {
		newValue := *v.Delta + *metric.Delta
		v.Delta = &newValue
		s.Metrics[key] = v
	} else {
		s.Metrics[key] = metric
	}
	s.counterKey[key] = true
}

func getKey(mType string, name string) string {
	return mType + name
}
