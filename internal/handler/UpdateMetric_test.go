package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func newUpdateRequest(metricType, metricName, metricValue, contentType string) *http.Request {
	url := "/update/" + metricType + "/" + metricName + "/" + metricValue
	req := httptest.NewRequest(http.MethodPost, url, nil)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return mux.SetURLVars(req, map[string]string{
		"type":  metricType,
		"name":  metricName,
		"value": metricValue,
	})
}

func TestUpdateMetricHandler_Gauge(t *testing.T) {
	t.Parallel()

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
			req := newUpdateRequest("gauge", tt.metricName, tt.metricValue, "text/plain")
			w := httptest.NewRecorder()
			UpdateMetricHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ожидался статус %d, получено %d (Body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestUpdateMetricHandler_Counter(t *testing.T) {
	t.Parallel()

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
			req := newUpdateRequest("counter", tt.metricName, tt.metricValue, "text/plain")
			w := httptest.NewRecorder()
			UpdateMetricHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ожидался статус %d, получено %d (Body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestUpdateMetricHandler_InvalidType(t *testing.T) {
	t.Parallel()

	req := newUpdateRequest("invalid", "test", "123", "text/plain")
	w := httptest.NewRecorder()
	UpdateMetricHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидался статус %d, получено %d (Body: %s)", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestUpdateMetricHandler_ContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		metricType  string
		metricValue string
		contentType string
		wantStatus  int
	}{
		{"text/plain", "gauge", "123.45", "text/plain", http.StatusOK},
		// хендлер сравнивает заголовок строго, поэтому параметры не допускаются
		{"text/plain с charset", "gauge", "123.45", "text/plain; charset=utf-8", http.StatusUnsupportedMediaType},
		{"application/json", "gauge", "123.45", "application/json", http.StatusUnsupportedMediaType},
		{"нет заголовка (gauge)", "gauge", "123.45", "", http.StatusUnsupportedMediaType},
		{"нет заголовка (counter)", "counter", "123", "", http.StatusUnsupportedMediaType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newUpdateRequest(tt.metricType, "test", tt.metricValue, tt.contentType)
			w := httptest.NewRecorder()
			UpdateMetricHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ожидался статус %d, получено %d (Body: %s)", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
