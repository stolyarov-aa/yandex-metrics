package handler

import (
	"fmt"
	"net/http"

	"github.com/stolyarov-aa/yandex-metrics/internal/repository"
)

func (h *HandlerHub) GetAllMetricsHandler(w http.ResponseWriter, r *http.Request) {
	var store repository.MemStorageIntf
	store = h.store

	counterMetrics := store.GetAllCounterMetrics()
	gaugeMetrics := store.GetAllGaugeMetrics()

	html := "<!DOCTYPE html><html><head><title>Yandex Metrics</title></head><body>"
	html = html + "<h1>Counter Metrics:</h1>"
	if len(counterMetrics) == 0 {
		html = html + "<p style=\"color:blue\">Метрики отсутствуют</p>"
	} else {
		for i, v := range counterMetrics {
			html = html + fmt.Sprintf("<p style=\"color:blue\">%q - %d</p>", i, v)
		}
	}
	html = html + "<h1>Gauge Metrics:</h1>"
	if len(gaugeMetrics) == 0 {
		html = html + "<p style=\"color:blue\">Метрики отсутствуют</p>"
	} else {
		for i, v := range gaugeMetrics {
			html = html + fmt.Sprintf("<p style=\"color:blue\">%q - %g</p>", i, v)
		}
	}
	html = html + "</body></html>"

	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, html)
}
