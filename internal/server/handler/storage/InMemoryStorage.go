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

func (s *InMemoryStorage) StoreGauge(metric data.Metric) {
	key := getKey(gauge, metric.ID)
	s.Metrics[key] = metric
	s.gaugeKey[key] = true
}

func (s *InMemoryStorage) GetCounter(name string) (metric data.Metric, exist bool) {
	metric, exist = s.Metrics[getKey(counter, name)]
	return metric, exist
}

func (s *InMemoryStorage) StoreCounter(metric data.Metric) {
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

func getKey(mType string, name string) string {
	return mType + name
}
