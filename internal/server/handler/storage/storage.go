package storage

type Storage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func (s *Storage) GetGauge(name string) (value float64, exist bool) {
	value, exist = s.Gauge[name]
	return value, exist
}

func (s *Storage) StoreGauge(name string, value float64) {
	s.Gauge[name] = value
}

func (s *Storage) GetCounter(name string) (value int64, exist bool) {
	value, exist = s.Counter[name]
	return value, exist
}

func (s *Storage) StoreCounter(name string, value int64) {
	v, ok := s.Counter[name]
	if ok {
		s.Counter[name] = v + value
	} else {
		s.Counter[name] = value
	}
}

func NewStorage() *Storage {
	return &Storage{Gauge: make(map[string]float64), Counter: map[string]int64{}}
}

func (s *Storage) GetGaugeMetrics() map[string]float64 {
	result := make(map[string]float64)
	for k, v := range s.Gauge {
		result[k] = v
	}
	return result
}

func (s *Storage) GetCounterMetrics() map[string]int64 {
	result := make(map[string]int64)
	for k, v := range s.Counter {
		result[k] = v
	}
	return result
}
