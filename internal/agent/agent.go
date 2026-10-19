package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/stolyarov-aa/yandex-metrics/internal/repository"
	"github.com/stolyarov-aa/yandex-metrics/internal/service"
	"github.com/stolyarov-aa/yandex-metrics/internal/storage"
)

type Agent struct {
	store          repository.MemStorageIntf
	serverAddr     string
	pollInterval   time.Duration
	reportInterval time.Duration
	httpClient     *http.Client
}

func New(serverAddr string, pollInterval int, reportInterval int) (*Agent, error) {
	if serverAddr == "" {
		serverAddr = "localhost:8080"
	}
	if pollInterval == 0 {
		pollInterval = 2
	}
	if reportInterval == 0 {
		reportInterval = 10
	}

	if err := service.ValidateHost(serverAddr); err != nil {
		return nil, err
	}

	if pollInterval <= 0 {
		return nil, fmt.Errorf("интервал pollInterval должен быть больше нуля")
	}

	if reportInterval <= 0 {
		return nil, fmt.Errorf("интервал reportInterval должен быть больше нуля")
	}

	a := &Agent{
		store:          storage.NewMemStorage(),
		serverAddr:     "http://" + serverAddr,
		pollInterval:   time.Duration(pollInterval) * time.Second,
		reportInterval: time.Duration(reportInterval) * time.Second,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	return a, nil
}

func (a *Agent) Run(ctx context.Context) {
	a.store.AddCounterMetric("PollCount", 0)
	a.store.SetGaugeMetric("RandomValue", rand.Float64())

	go a.startPolling(ctx)

	ticker := time.NewTicker(a.reportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("Агент остановлен")
			return
		case <-ticker.C:
			a.send()
		}
	}
}

func (a *Agent) startPolling(ctx context.Context) {
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.getMetrics()
		}
	}
}

func (a *Agent) getMetrics() {
	rnd := rand.Float64()
	a.store.SetGaugeMetric("RandomValue", rnd)
	for _, m := range service.GetRuntimeMetrics() {
		a.store.SetGaugeMetric(m.Name, m.Value)
	}
	a.store.AddCounterMetric("PollCount", 1)
}

func (a *Agent) send() {
	fnSend := func(metricType, metricName, metricValue string) {
		statusCode, err := a.sendMetric(metricType, metricName, metricValue)
		if err != nil {
			log.Printf("Ошибка при отправке %s-метрики %q: %v", metricType, metricName, err)
		} else {
			log.Printf("%s-метрика %q отправлена, статус: %d", metricType, metricName, statusCode)
		}
	}

	for name, value := range a.store.GetAllGaugeMetrics() {
		fnSend("gauge", name, strconv.FormatFloat(value, 'g', 10, 64))
	}
	for name, value := range a.store.GetAllCounterMetrics() {
		fnSend("counter", name, strconv.FormatInt(value, 10))
	}
}

func (a *Agent) sendMetric(metricType, metricName, metricValue string) (int, error) {
	url := fmt.Sprintf("%s/update/%s/%s/%s", a.serverAddr, metricType, metricName, metricValue)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
