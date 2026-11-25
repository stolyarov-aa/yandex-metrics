package storage

import (
	"sync"
	"testing"
)

func TestMemStorage_SetGaugeMetric(t *testing.T) {
	t.Parallel()

	s := NewMemStorage()

	s.SetGaugeMetric("test", 123.45)
	val, ok := s.GetGaugeMetric("test")
	if !ok {
		t.Errorf("метрика не найдена")
	}
	if val != 123.45 {
		t.Errorf("ожидалось 123.45, получено %f", val)
	}

	s.SetGaugeMetric("test", 456.78)
	val, _ = s.GetGaugeMetric("test")
	if val != 456.78 {
		t.Errorf("ожидалось 456.78, получено %f", val)
	}
}

func TestMemStorage_AddCounterMetric(t *testing.T) {
	t.Parallel()

	s := NewMemStorage()

	s.AddCounterMetric("test", 10)
	val, ok := s.GetCounterMetric("test")
	if !ok {
		t.Errorf("метрика не найдена")
	}
	if val != 10 {
		t.Errorf("ожидалось 10, получено %d", val)
	}

	s.AddCounterMetric("test", 5)
	val, _ = s.GetCounterMetric("test")
	if val != 15 {
		t.Errorf("ожидалось 15, получено %d", val)
	}
}

func TestMemStorage_GetAllGaugeMetrics(t *testing.T) {
	t.Parallel()

	s := NewMemStorage()
	s.SetGaugeMetric("a", 1.1)
	s.SetGaugeMetric("b", 2.2)

	metrics := s.GetAllGaugeMetrics()
	if len(metrics) != 2 {
		t.Errorf("ожидалось 2 метрики, получено %d", len(metrics))
	}
	if metrics["a"] != 1.1 {
		t.Errorf("ожидалось 1.1 для 'a', получено %f", metrics["a"])
	}
	if metrics["b"] != 2.2 {
		t.Errorf("ожидалось 2.2 для 'b', получено %f", metrics["b"])
	}
}

func TestMemStorage_GetAllCounterMetrics(t *testing.T) {
	t.Parallel()

	s := NewMemStorage()
	s.AddCounterMetric("x", 100)
	s.AddCounterMetric("y", 200)

	metrics := s.GetAllCounterMetrics()
	if len(metrics) != 2 {
		t.Errorf("ожидалось 2 метрики, получено %d", len(metrics))
	}
	if metrics["x"] != 100 {
		t.Errorf("ожидалось 100 для 'x', получено %d", metrics["x"])
	}
	if metrics["y"] != 200 {
		t.Errorf("ожидалось 200 для 'y', получено %d", metrics["y"])
	}
}

func TestMemStorage_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	s := NewMemStorage()
	var wg sync.WaitGroup

	// Горутины для записи
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.SetGaugeMetric("gauge", float64(i))
			s.AddCounterMetric("counter", 1)
		}(i)
	}

	wg.Wait()

	if _, ok := s.GetGaugeMetric("gauge"); !ok {
		t.Errorf("метрика gauge не найдена")
	}
	if _, ok := s.GetCounterMetric("counter"); !ok {
		t.Errorf("метрика counter не найдена")
	}
}
