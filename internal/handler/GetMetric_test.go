package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stolyarov-aa/yandex-metrics/internal/storage"
)

func newGetTestRouter() *chi.Mux {
	r := chi.NewRouter()

	store := storage.NewMemStorage()
	hub := &HandlerHub{
		store: store,
	}

	r.Post("/update/{type}/{name}/{value}", func(w http.ResponseWriter, r *http.Request) {
		hub.UpdateMetricHandler(w, r)
	})

	r.Get("/value/{type}/{name}", func(w http.ResponseWriter, r *http.Request) {
		hub.GetMetricHandler(w, r)
	})

	return r
}

func TestGetMetricHandler_Gauge(t *testing.T) {
	t.Parallel()

	router := newGetTestRouter()

	tests := []struct {
		name       string
		metricName string
		valueToSet string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "получение существующей gauge",
			metricName: "testGauge",
			valueToSet: "123.45",
			wantStatus: http.StatusOK,
			wantBody:   "123.45",
		},
		{
			name:       "получение gauge с целым значением",
			metricName: "testGaugeInt",
			valueToSet: "100",
			wantStatus: http.StatusOK,
			wantBody:   "100",
		},
		{
			name:       "получение несуществующей gauge",
			metricName: "unknownGauge",
			valueToSet: "",
			wantStatus: http.StatusNotFound,
			wantBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valueToSet != "" {
				updateURL := "/update/gauge/" + tt.metricName + "/" + tt.valueToSet
				updateReq := httptest.NewRequest(http.MethodPost, updateURL, nil)
				updateReq.Header.Set("Content-Type", "text/plain")
				updateW := httptest.NewRecorder()
				router.ServeHTTP(updateW, updateReq)

				if updateW.Code != http.StatusOK {
					t.Fatalf("Не удалось записать метрику: статус %d, тело %s", updateW.Code, updateW.Body.String())
				}
			}

			getURL := "/value/gauge/" + tt.metricName
			getReq := httptest.NewRequest(http.MethodGet, getURL, nil)
			getW := httptest.NewRecorder()
			router.ServeHTTP(getW, getReq)

			if getW.Code != tt.wantStatus {
				t.Errorf("Ожидался статус %d, получено %d (Body: %s)", tt.wantStatus, getW.Code, getW.Body.String())
			}

			if tt.wantStatus == http.StatusOK && getW.Body.String() != tt.wantBody {
				t.Errorf("Ожидалось тело %q, получено %q", tt.wantBody, getW.Body.String())
			}
		})
	}
}

func TestGetMetricHandler_Counter(t *testing.T) {
	t.Parallel()

	router := newGetTestRouter()

	tests := []struct {
		name        string
		metricName  string
		valuesToSet []string
		wantStatus  int
		wantBody    string
	}{
		{
			name:        "получение существующего counter (один инкремент)",
			metricName:  "testCounter1",
			valuesToSet: []string{"10"},
			wantStatus:  http.StatusOK,
			wantBody:    "10",
		},
		{
			name:        "получение существующего counter (несколько инкрементов)",
			metricName:  "testCounter2",
			valuesToSet: []string{"5", "10", "5"}, // 5+10+5=20
			wantStatus:  http.StatusOK,
			wantBody:    "20",
		},
		{
			name:        "получение несуществующего counter",
			metricName:  "unknownCounter",
			valuesToSet: nil,
			wantStatus:  http.StatusNotFound,
			wantBody:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, val := range tt.valuesToSet {
				updateURL := "/update/counter/" + tt.metricName + "/" + val
				updateReq := httptest.NewRequest(http.MethodPost, updateURL, nil)
				updateReq.Header.Set("Content-Type", "text/plain")
				updateW := httptest.NewRecorder()
				router.ServeHTTP(updateW, updateReq)

				if updateW.Code != http.StatusOK {
					t.Fatalf("Не удалось обновить counter: статус %d, тело %s", updateW.Code, updateW.Body.String())
				}
			}

			getURL := "/value/counter/" + tt.metricName
			getReq := httptest.NewRequest(http.MethodGet, getURL, nil)
			getW := httptest.NewRecorder()
			router.ServeHTTP(getW, getReq)

			if getW.Code != tt.wantStatus {
				t.Errorf("Ожидался статус %d, получено %d (Body: %s)", tt.wantStatus, getW.Code, getW.Body.String())
			}

			if tt.wantStatus == http.StatusOK && getW.Body.String() != tt.wantBody {
				t.Errorf("Ожидалось тело %q, получено %q", tt.wantBody, getW.Body.String())
			}
		})
	}
}

func TestGetMetricHandler_InvalidType(t *testing.T) {
	t.Parallel()

	router := newGetTestRouter()

	getURL := "/value/unknown/testMetric"
	getReq := httptest.NewRequest(http.MethodGet, getURL, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Errorf("Ожидался статус %d, получено %d (Body: %s)", http.StatusNotFound, getW.Code, getW.Body.String())
	}
}
