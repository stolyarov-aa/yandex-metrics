package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stolyarov-aa/yandex-metrics/internal/storage"
)

func newTestRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
		hub := &HandlerHub{
			store: storage.NewMemStorage(),
		}
		hub.UpdateMetricHandler(w, r)
	})
	return r
}

func TestUpdateMetricHandler_Gauge(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	tests := []struct {
		name        string
		metricName  string
		metricValue string
		wantStatus  int
	}{
		{"валидная gauge (float)", "test", "123.45", http.StatusOK},
		{"валидная gauge (int)", "test", "123", http.StatusOK},
		{"валидная gauge (отрицательная)", "test", "-1.5", http.StatusOK},
		{"пустое имя", "", "123", http.StatusNotFound},
		{"неверное значение", "test", "abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/update/gauge/" + tt.metricName + "/" + tt.metricValue
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ожидался статус %d, получено %d (Body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestUpdateMetricHandler_Counter(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	tests := []struct {
		name        string
		metricName  string
		metricValue string
		wantStatus  int
	}{
		{"валидный counter", "test", "123", http.StatusOK},
		{"counter с дробным значением", "test", "123.0", http.StatusBadRequest},
		{"пустое имя", "", "123", http.StatusNotFound},
		{"неверное значение", "test", "abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/update/counter/" + tt.metricName + "/" + tt.metricValue
			req := httptest.NewRequest(http.MethodPost, url, nil)
			req.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ожидался статус %d, получено %d (Body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestUpdateMetricHandler_InvalidType(t *testing.T) {
	t.Parallel()

	router := newTestRouter()

	url := "/update/invalid/test/123"
	req := httptest.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидался статус %d, получено %d (Body: %s)", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

// func TestUpdateMetricHandler_ContentType(t *testing.T) {
// 	t.Parallel()

// 	router := newTestRouter()

// 	tests := []struct {
// 		name        string
// 		metricType  string
// 		metricValue string
// 		contentType string
// 		wantStatus  int
// 	}{
// 		{"text/plain", "gauge", "123.45", "text/plain", http.StatusOK},
// 		{"text/plain с charset", "gauge", "123.45", "text/plain; charset=utf-8", http.StatusUnsupportedMediaType},
// 		{"application/json", "gauge", "123.45", "application/json", http.StatusUnsupportedMediaType},
// 		{"нет заголовка (gauge)", "gauge", "123.45", "", http.StatusUnsupportedMediaType},
// 		{"нет заголовка (counter)", "counter", "123", "", http.StatusUnsupportedMediaType},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			url := "/update/" + tt.metricType + "/test/" + tt.metricValue
// 			req := httptest.NewRequest(http.MethodPost, url, nil)
// 			if tt.contentType != "" {
// 				req.Header.Set("Content-Type", tt.contentType)
// 			}

// 			w := httptest.NewRecorder()
// 			router.ServeHTTP(w, req)

// 			if w.Code != tt.wantStatus {
// 				t.Errorf("ожидался статус %d, получено %d (Body: %s)", tt.wantStatus, w.Code, w.Body.String())
// 			}
// 		})
// 	}
// }
