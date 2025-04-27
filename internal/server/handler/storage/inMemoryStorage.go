package storage

import (
	"context"

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

func (s *InMemoryStorage) GetGauge(name string) (metric data.Metric, exist bool, err error) {
	metric, exist = s.Metrics[getKey(gauge, name)]
	return metric, exist, nil
}

func (s *InMemoryStorage) Store(ctx context.Context, metric ...data.Metric) error {
	for _, v := range metric {
		key := getKey(v.MType, v.ID)
		if v.MType == gauge {
			s.Metrics[key] = v
			s.gaugeKey[key] = true
		} else {
			s.storeCounter(v)
		}
	}
	return nil
}

func (s *InMemoryStorage) GetCounter(name string) (metric data.Metric, exist bool, err error) {
	metric, exist = s.Metrics[getKey(counter, name)]
	return metric, exist, nil
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{Metrics: map[string]data.Metric{}, gaugeKey: map[string]bool{}, counterKey: map[string]bool{}}
}

func (s *InMemoryStorage) GetMetrics() ([]data.Metric, error) {
	result := make([]data.Metric, 0)
	for k := range s.counterKey {
		m, ok := s.Metrics[k]
		if ok {
			result = append(result, m)
		}

	}
	for k := range s.gaugeKey {
		m, ok := s.Metrics[k]
		if ok {
			result = append(result, m)
		}

	}
	return result, nil
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
