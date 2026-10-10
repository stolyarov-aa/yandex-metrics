package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// Ошибки для метрик
type (
	InvalidMetricTypeError struct {
		MetricType string
	}
	InvalidMetricNameError struct {
		MetricName string
	}
	InvalidMetricDataTypeError struct {
		MetricType string
		MetricName string
		DataType   string
	}
)

func (e *InvalidMetricTypeError) Error() string {
	return fmt.Sprintf("недопустимый тип метрики %q: разрешены только 'gauge' и 'counter'", e.MetricType)
}
func (e *InvalidMetricTypeError) As(target any) bool {
	if target == nil {
		return false
	}
	if t, ok := target.(**InvalidMetricTypeError); ok {
		*t = e
		return true
	}
	return false
}

func (e *InvalidMetricNameError) Error() string {
	return fmt.Sprintf("недопустимое имя метрики %q", e.MetricName)
}
func (e *InvalidMetricNameError) As(target any) bool {
	if t, ok := target.(**InvalidMetricNameError); ok {
		*t = e
		return true
	}
	return false
}

func (e *InvalidMetricDataTypeError) Error() string {
	return fmt.Sprintf("неверный тип данных %s-метрики %q: %s", e.MetricType, e.MetricName, e.DataType)
}
func (e *InvalidMetricDataTypeError) As(target any) bool {
	if t, ok := target.(**InvalidMetricDataTypeError); ok {
		*t = e
		return true
	}
	return false
}

// Хранилище и действия с ними
type (
	MemStorage struct {
		mu      sync.RWMutex
		Gauge   map[string]float64
		Counter map[string]int64
	}
)

func (s *MemStorage) setGauge(metricName string, metricValue float64) {
	s.Gauge[metricName] = metricValue
}

func (s *MemStorage) incCounter(metricName string, metricValue int64) {
	s.Counter[metricName] = s.Counter[metricName] + metricValue
}

func (s *MemStorage) Update(metricType string, metricName string, metricValue interface{}) error {
	if metricType != "gauge" && metricType != "counter" {
		return &InvalidMetricTypeError{MetricType: metricType}
	} else if metricName == "" {
		return &InvalidMetricNameError{MetricName: "<<Имя не указано>>"}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch v := metricValue.(type) {
	case float64:
		if metricType != "gauge" {
			return &InvalidMetricDataTypeError{MetricType: metricType, DataType: fmt.Sprintf("%T", v)}
		}
		s.setGauge(metricName, v)
	case int64:
		if metricType != "counter" {
			return &InvalidMetricDataTypeError{MetricType: metricType, DataType: fmt.Sprintf("%T", v)}
		}
		s.incCounter(metricName, v)
	case string:
		switch metricType {
		case "gauge":
			if n, err := strconv.ParseFloat(v, 64); err != nil {
				return &InvalidMetricDataTypeError{MetricType: metricType, MetricName: metricName, DataType: fmt.Sprintf("%T", v)}
			} else {
				s.Gauge[metricName] = n
			}
		case "counter":
			if n, err := strconv.ParseInt(v, 10, 64); err != nil {
				return &InvalidMetricDataTypeError{MetricType: metricType, MetricName: metricName, DataType: fmt.Sprintf("%T", v)}
			} else {
				s.Counter[metricName] = s.Counter[metricName] + n
			}
		default:
			return &InvalidMetricDataTypeError{MetricType: metricType, MetricName: metricName, DataType: fmt.Sprintf("%T", v)}
		}
	default:
		return &InvalidMetricDataTypeError{MetricType: metricType, MetricName: metricName, DataType: fmt.Sprintf("%T", v)}
	}

	return nil
}

func (s *MemStorage) Get(metricType string, metricName string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch metricType {
	case "gauge":
		return s.Gauge[metricName], true
	case "counter":
		return s.Counter[metricName], true
	default:
		return 0, false
	}
}

// Сервис для обновления метрик
type MemStorageInf interface {
	Update(metricType string, metricName string, metricValue interface{}) error
	Get(metricType string, metricName string) (interface{}, bool)
}

var storage MemStorageInf = &MemStorage{
	Gauge:   make(map[string]float64),
	Counter: make(map[string]int64),
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if str := r.Header.Get("Content-Type"); str != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusUnsupportedMediaType)
		return
	}

	vars := mux.Vars(r)
	metricType := vars["type"]
	metricName := vars["name"]
	metricValue := vars["value"]

	if err := storage.Update(metricType, metricName, metricValue); err != nil {
		switch err.(type) {
		case *InvalidMetricTypeError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case *InvalidMetricNameError:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *InvalidMetricDataTypeError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	r := mux.NewRouter()
	r.StrictSlash(false)
	r.HandleFunc("/update/{type}/{name}/{value}", updateHandler).Methods("POST")

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}