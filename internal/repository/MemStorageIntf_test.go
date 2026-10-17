package repository

import (
	"testing"

	"github.com/stolyarov-aa/yandex-metrics/internal/storage"
)

func TestMemStorageIntf_SetGaugeMetric(t *testing.T) {
	var s MemStorageIntf = storage.NewMemStorage()
	s.SetGaugeMetric("test", 123.45)
	val, ok := s.GetGaugeMetric("test")
	if !ok || val != 123.45 {
		t.Errorf("ожидалось 123.45, получено %f, ok=%v", val, ok)
	}
}

func TestMemStorageIntf_AddCounterMetric(t *testing.T) {
	var s MemStorageIntf = storage.NewMemStorage()
	s.AddCounterMetric("test", 10)
	val, ok := s.GetCounterMetric("test")
	if !ok || val != 10 {
		t.Errorf("ожидалось 10, получено %d, ok=%v", val, ok)
	}
}
