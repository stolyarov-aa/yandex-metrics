package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testServer — тестовый HTTP-сервер, потокобезопасно запоминающий пути запросов.
type testServer struct {
	*httptest.Server
	mu    sync.Mutex
	paths []string
}

func newTestServer(t *testing.T, status int) *testServer {
	t.Helper()
	ts := &testServer{}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.paths = append(ts.paths, r.URL.Path)
		ts.mu.Unlock()
		w.WriteHeader(status)
	}))
	t.Cleanup(ts.Close)
	return ts
}

// Paths возвращает копию путей всех полученных запросов.
func (ts *testServer) Paths() []string {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return append([]string(nil), ts.paths...)
}

// CountPrefix считает запросы, путь которых начинается с prefix.
func (ts *testServer) CountPrefix(prefix string) int {
	n := 0
	for _, p := range ts.Paths() {
		if strings.HasPrefix(p, prefix) {
			n++
		}
	}
	return n
}

func TestNew(t *testing.T) {
	type want struct {
		panic bool
	}
	tests := []struct {
		name           string
		serverAddr     string
		pollInterval   time.Duration
		reportInterval time.Duration
		want           want
	}{
		{
			name:           "пустой serverAddr",
			serverAddr:     "",
			pollInterval:   time.Second,
			reportInterval: time.Second,
			want:           want{panic: true},
		},
		{
			name:           "нулевой pollInterval",
			serverAddr:     "http://localhost:8080",
			pollInterval:   0,
			reportInterval: time.Second,
			want:           want{panic: true},
		},
		{
			name:           "нулевой reportInterval",
			serverAddr:     "http://localhost:8080",
			pollInterval:   time.Second,
			reportInterval: 0,
			want:           want{panic: true},
		},
		{
			name:           "валидные параметры",
			serverAddr:     "http://localhost:8080",
			pollInterval:   2 * time.Second,
			reportInterval: 10 * time.Second,
			want:           want{panic: false},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			create := func() {
				New(test.serverAddr, test.pollInterval, test.reportInterval)
			}

			if test.want.panic {
				assert.Panics(t, create)
				return
			}
			require.NotPanics(t, create)

			a := New(test.serverAddr, test.pollInterval, test.reportInterval)
			assert.Equal(t, test.serverAddr, a.serverAddr)
			assert.Equal(t, test.pollInterval, a.pollInterval)
			assert.Equal(t, test.reportInterval, a.reportInterval)
			assert.NotNil(t, a.store)
			require.NotNil(t, a.httpClient)
			assert.NotZero(t, a.httpClient.Timeout)
		})
	}
}

func TestGetMetrics(t *testing.T) {
	type want struct {
		pollCount int64
	}
	tests := []struct {
		name  string
		polls int
		want  want
	}{
		{
			name:  "один опрос",
			polls: 1,
			want:  want{pollCount: 1},
		},
		{
			name:  "три опроса",
			polls: 3,
			want:  want{pollCount: 3},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := New("http://localhost:8080", time.Second, time.Second)

			for i := 0; i < test.polls; i++ {
				a.getMetrics()
			}

			count, ok := a.store.GetCounterMetric("PollCount")
			require.True(t, ok, "нет метрики PollCount")
			assert.Equal(t, test.want.pollCount, count)

			rnd, ok := a.store.GetGaugeMetric("RandomValue")
			require.True(t, ok, "нет метрики RandomValue")
			assert.GreaterOrEqual(t, rnd, 0.0)
			assert.Less(t, rnd, 1.0) // rand.Float64() возвращает [0, 1)

			heapAlloc, ok := a.store.GetGaugeMetric("HeapAlloc")
			require.True(t, ok, "нет метрики HeapAlloc")
			assert.Positive(t, heapAlloc)
		})
	}
}

func TestSend(t *testing.T) {
	type want struct {
		gauges   int
		counters int
		paths    []string
	}
	tests := []struct {
		name     string
		gauges   map[string]float64
		counters map[string]int64
		want     want
	}{
		{
			name:     "gauge и counter",
			gauges:   map[string]float64{"TestGauge": 42},
			counters: map[string]int64{"TestCounter": 10},
			want: want{
				gauges:   1,
				counters: 1,
				paths:    []string{"/update/gauge/TestGauge/42", "/update/counter/TestCounter/10"},
			},
		},
		{
			name:   "дробный gauge",
			gauges: map[string]float64{"Frac": 0.125},
			want: want{
				gauges: 1,
				paths:  []string{"/update/gauge/Frac/0.125"},
			},
		},
		{
			name: "пустое хранилище",
			want: want{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := newTestServer(t, http.StatusOK)
			a := New(srv.URL, time.Second, time.Second)

			for name, v := range test.gauges {
				require.NoError(t, a.store.SetGaugeMetric(name, v))
			}
			for name, v := range test.counters {
				require.NoError(t, a.store.AddCounterMetric(name, v))
			}

			// send() синхронный: после возврата все запросы уже обработаны.
			a.send()

			assert.Equal(t, test.want.gauges, srv.CountPrefix("/update/gauge/"))
			assert.Equal(t, test.want.counters, srv.CountPrefix("/update/counter/"))
			assert.ElementsMatch(t, test.want.paths, srv.Paths())
		})
	}
}

func TestRun(t *testing.T) {
	type want struct {
		requests bool
	}
	tests := []struct {
		name           string
		serverDown     bool
		pollInterval   time.Duration
		reportInterval time.Duration
		want           want
	}{
		{
			name:           "отправляет метрики и останавливается",
			pollInterval:   20 * time.Millisecond,
			reportInterval: 100 * time.Millisecond,
			want:           want{requests: true},
		},
		{
			name:           "останавливается при недоступном сервере",
			serverDown:     true,
			pollInterval:   20 * time.Millisecond,
			reportInterval: 50 * time.Millisecond,
			want:           want{requests: false},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			srv := newTestServer(t, http.StatusOK)
			if test.serverDown {
				srv.Close()
			}
			a := New(srv.URL, test.pollInterval, test.reportInterval)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			done := make(chan struct{})
			go func() {
				a.Run(ctx)
				close(done)
			}()

			if test.want.requests {
				require.Eventually(t, func() bool {
					return srv.CountPrefix("/update/counter/PollCount/") > 0 &&
						srv.CountPrefix("/update/gauge/RandomValue/") > 0 &&
						srv.CountPrefix("/update/gauge/HeapAlloc/") > 0
				}, 2*time.Second, 10*time.Millisecond, "агент не отправил метрики")
			} else {
				time.Sleep(3 * test.reportInterval)
			}

			cancel()

			select {
			case <-done:
			case <-time.After(2 * time.Second):
				require.Fail(t, "агент не остановился после отмены контекста")
			}
		})
	}
}
