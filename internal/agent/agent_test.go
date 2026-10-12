package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serverAddr     string
		pollInterval   time.Duration
		reportInterval time.Duration
		wantPanic      bool
	}{
		{"пустой serverAddr", "", 1 * time.Second, 1 * time.Second, true},
		{"ноль pollInterval", "http://localhost:8080", 0, 1 * time.Second, true},
		{"ноль reportInterval", "http://localhost:8080", 1 * time.Second, 0, true},
		{"валидные параметры", "http://localhost:8080", 1 * time.Second, 2 * time.Second, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil && tt.wantPanic {
					t.Errorf("ожидалась паника, но её не было")
				} else if r != nil && !tt.wantPanic {
					t.Errorf("неожиданная паника: %v", r)
				}
			}()

			a := New(tt.serverAddr, tt.pollInterval, tt.reportInterval)
			if a == nil && !tt.wantPanic {
				t.Errorf("ожидался Agent, получено nil")
			}
		})
	}
}

func TestGetMetrics(t *testing.T) {
	t.Parallel()

	a := New("http://localhost:9999", time.Second, time.Second)

	a.getMetrics()

	val, ok := a.store.GetGaugeMetric("RandomValue")
	if !ok {
		t.Error("ожидалась метрика RandomValue")
	}
	if val < 0 || val > 1 {
		t.Errorf("RandomValue должен быть в диапазоне [0, 1], получено %f", val)
	}

	count, ok := a.store.GetCounterMetric("PollCount")
	if !ok {
		t.Error("ожидалась метрика PollCount")
	}
	if count != 1 {
		t.Errorf("ожидался PollCount=1, получено %d", count)
	}

	heapAlloc, ok := a.store.GetGaugeMetric("HeapAlloc")
	if !ok {
		t.Error("ожидалась метрика HeapAlloc")
	}
	if heapAlloc <= 0 {
		t.Errorf("HeapAlloc должен быть > 0, получено %f", heapAlloc)
	}
}

func TestSend(t *testing.T) {
	t.Parallel()

	var receivedGauges int
	var receivedCounters int

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) < 9 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if path[9:14] == "gauge" {
			receivedGauges++
		} else if path[9:16] == "counter" {
			receivedCounters++
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	a := New(testServer.URL, time.Second, time.Second)

	// Добавляем тестовые метрики
	a.store.SetGaugeMetric("TestGauge", 42.0)
	a.store.AddCounterMetric("TestCounter", 10)

	a.send()

	if receivedGauges == 0 {
		t.Error("ожидалась отправка gauge-метрик")
	}
	if receivedCounters == 0 {
		t.Error("ожидалась отправка counter-метрик")
	}
}

func TestSendMetric_Success(t *testing.T) {
	t.Parallel()

	var receivedPath string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	a := New(testServer.URL, time.Second, time.Second)

	statusCode, err := a.sendMetric("gauge", "TestMetric", "42")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if statusCode != http.StatusOK {
		t.Errorf("ожидался статус 200, получено %d", statusCode)
	}
	if receivedPath != "/update/gauge/TestMetric/42" {
		t.Errorf("ожидался путь /update/gauge/TestMetric/42, получено %s", receivedPath)
	}
}

func TestSendMetric_Error(t *testing.T) {
	t.Parallel()

	a := New("http://localhost:1", time.Second, time.Second)

	_, err := a.sendMetric("gauge", "TestMetric", "42")
	if err == nil {
		t.Error("ожидалась ошибка при отправке на недоступный сервер")
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	t.Parallel()

	a := New("http://localhost:1", 50*time.Millisecond, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		a.Run(ctx)
		close(done)
	}()

	time.Sleep(150 * time.Millisecond)

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("агент не остановился после отмены контекста")
	}
}

func TestRun_SendsMetrics(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var receivedRequests []string

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedRequests = append(receivedRequests, r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	a := New(testServer.URL, 50*time.Millisecond, 100*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		a.Run(ctx)
		close(done)
	}()

	time.Sleep(300 * time.Millisecond)
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()

	if len(receivedRequests) == 0 {
		t.Error("ожидались отправленные метрики")
	}
}
