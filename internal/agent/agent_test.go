package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestHandleUpdate_Gauge(t *testing.T) {
	a := New("http://localhost:9999", time.Second, time.Second)

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"валидный gauge", "/update/gauge/testGauge/100", http.StatusOK},
		{"без id", "/update/gauge/", http.StatusNotFound},
		{"некорректный value", "/update/gauge/testGauge/none", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, nil)
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			a.handleUpdate(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("ожидался статус %d, получено %d", tt.wantCode, w.Code)
			}
		})
	}
}

func TestHandleUpdate_Counter(t *testing.T) {
	a := New("http://localhost:9999", time.Second, time.Second)

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"валидный counter", "/update/counter/testCounter/100", http.StatusOK},
		{"без id", "/update/counter/", http.StatusNotFound},
		{"некорректный value", "/update/counter/testCounter/none", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, nil)
			req.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()

			a.handleUpdate(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("ожидался статус %d, получено %d", tt.wantCode, w.Code)
			}
		})
	}
}

func TestHandleUpdate_UnknownType(t *testing.T) {
	a := New("http://localhost:9999", time.Second, time.Second)

	req := httptest.NewRequest("POST", "/update/unknown/testCounter/100", nil)
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	a.handleUpdate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидался статус %d, получено %d", http.StatusBadRequest, w.Code)
	}
}

func TestRun_ListensOnPort(t *testing.T) {
	a := New("http://localhost:9999", 100*time.Millisecond, 200*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.Run(ctx)

	time.Sleep(200 * time.Millisecond)

	resp, err := http.Post("http://localhost:8080/update/gauge/test/42", "text/plain", nil)
	if err != nil {
		t.Fatalf("не удалось сделать запрос: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("ожидался статус 200, получено %d", resp.StatusCode)
	}
}
