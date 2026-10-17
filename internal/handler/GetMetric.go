package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *HandlerHub) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	store := h.store

	switch metricType {
	case "counter":
		if v, ok := store.GetCounterMetric(metricName); ok {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%d", v)
			return
		}
	case "gauge":
		if v, ok := store.GetGaugeMetric(metricName); ok {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%g", v)
			return
		}
	default:
	}
	w.WriteHeader(http.StatusNotFound)
}
