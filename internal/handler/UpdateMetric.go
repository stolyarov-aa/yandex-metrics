package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	errs "github.com/stolyarov-aa/yandex-metrics/internal/errors"
	"github.com/stolyarov-aa/yandex-metrics/internal/repository"
)

func updateMetric(h *HandlerHub, metricType string, metricName string, metricValue interface{}) error {
	var s repository.MemStorageIntf = h.store

	if metricType != "gauge" && metricType != "counter" {
		return &errs.InvalidMetricTypeError{MetricType: metricType}
	} else if metricName == "" {
		return &errs.InvalidMetricNameError{MetricName: "<<Имя не указано>>"}
	}

	switch v := metricValue.(type) {
	case float64:
		if metricType != "gauge" {
			return &errs.InvalidMetricDataTypeError{MetricType: metricType, DataType: fmt.Sprintf("%T", v)}
		}
		s.SetGaugeMetric(metricName, v)
	case int64:
		if metricType != "counter" {
			return &errs.InvalidMetricDataTypeError{MetricType: metricType, DataType: fmt.Sprintf("%T", v)}
		}
		s.AddCounterMetric(metricName, v)
	case string:
		switch metricType {
		case "gauge":
			if n, err := strconv.ParseFloat(v, 64); err != nil {
				return &errs.InvalidMetricDataTypeError{MetricType: metricType, MetricName: metricName, DataType: fmt.Sprintf("%T", v)}
			} else {
				s.SetGaugeMetric(metricName, n)
			}
		case "counter":
			if n, err := strconv.ParseInt(v, 10, 64); err != nil {
				return &errs.InvalidMetricDataTypeError{MetricType: metricType, MetricName: metricName, DataType: fmt.Sprintf("%T", v)}
			} else {
				s.AddCounterMetric(metricName, n)
			}
		default:
			return &errs.InvalidMetricDataTypeError{MetricType: metricType, MetricName: metricName, DataType: fmt.Sprintf("%T", v)}
		}
	default:
		return &errs.InvalidMetricDataTypeError{MetricType: metricType, MetricName: metricName, DataType: fmt.Sprintf("%T", v)}
	}

	return nil
}

func (h *HandlerHub) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	// if str := r.Header.Get("Content-Type"); str != "text/plain" {
	// 	http.Error(w, "Content-Type должен быть text/plain", http.StatusUnsupportedMediaType)
	// 	return
	// }

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if err := updateMetric(h, metricType, metricName, metricValue); err != nil {
		switch err.(type) {
		case *errs.InvalidMetricTypeError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case *errs.InvalidMetricNameError:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *errs.InvalidMetricDataTypeError:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
