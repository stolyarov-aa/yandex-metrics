package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stolyarov-aa/yandex-metrics/internal/storage"
)

func newTestRouterGetAll() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		hub := &HandlerHub{
			store: storage.NewMemStorage(),
		}
		hub.GetAllMetricsHandler(w, r)
	})
	return r
}

func TestGetAllMetricsHandler_EmptyStorage(t *testing.T) {
	t.Parallel()

	router := newTestRouterGetAll()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус %d, получено %d (Body: %s)", http.StatusOK, w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "Метрики отсутствуют") {
		t.Errorf("ожидалось сообщение 'Метрики отсутствуют' в HTML, получено: %s", body)
	}
}

func TestGetAllMetricsHandler_WithMetrics(t *testing.T) {
	t.Parallel()

	store := storage.NewMemStorage()

	fullRouter := chi.NewRouter()
	fullRouter.Post("/update/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
		hub := &HandlerHub{
			store: store,
		}
		hub.UpdateMetricHandler(w, r)
	})
	fullRouter.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		hub := &HandlerHub{
			store: store,
		}
		hub.GetAllMetricsHandler(w, r)
	})

	tests := []struct {
		metricType  string
		metricName  string
		metricValue string
	}{
		{"gauge", "gauge1", "123.45"},
		{"counter", "counter1", "100"},
		{"gauge", "gauge2", "-50.1"},
		{"counter", "counter2", "200"},
	}

	for _, tt := range tests {
		url := "/update/" + tt.metricType + "/" + tt.metricName + "/" + tt.metricValue
		req := httptest.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		fullRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("не удалось добавить метрику %s/%s: статус %d", tt.metricType, tt.metricName, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	fullRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус %d, получено %d (Body: %s)", http.StatusOK, w.Code, w.Body.String())
	}

	body := w.Body.String()

	expectedStrings := []string{
		"\"gauge1\" - 123.45",
		"\"counter1\" - 100",
		"\"gauge2\" - -50.1",
		"\"counter2\" - 200",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(body, s) {
			t.Errorf("ожидалось наличие '%s' в HTML, получено: %s", s, body)
		}
	}
}

func TestGetAllMetricsHandler_ContentType(t *testing.T) {
	t.Parallel()

	router := newTestRouterGetAll()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус %d, получено %d", http.StatusOK, w.Code)
	}

	if w.Body.Len() == 0 {
		t.Error("тело ответа пустое")
	}
}
