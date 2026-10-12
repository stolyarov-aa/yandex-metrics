package storage

import (
	"sync"
)

type MemStorage struct {
	mu      sync.RWMutex
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (s *MemStorage) SetGaugeMetric(name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Gauge[name] = value
	return nil
}

func (s *MemStorage) GetGaugeMetric(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.Gauge[name]
	return value, ok
}

func (s *MemStorage) GetAllGaugeMetrics() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	retVal := make(map[string]float64)
	for k, v := range s.Gauge {
		retVal[k] = v
	}
	return retVal
}

func (s *MemStorage) AddCounterMetric(name string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if val, ok := s.Counter[name]; ok {
		s.Counter[name] = val + value
	} else {
		s.Counter[name] = value
	}
	return nil
}

func (s *MemStorage) GetCounterMetric(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.Counter[name]
	return value, ok
}

func (s *MemStorage) GetAllCounterMetrics() map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	retVal := make(map[string]int64)
	for k, v := range s.Counter {
		retVal[k] = v
	}
	return retVal
}
