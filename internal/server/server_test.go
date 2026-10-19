package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doRequest(t *testing.T, h http.Handler, method, path string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res.StatusCode, string(body)
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	s, err := New("")
	require.NoError(t, err)
	require.NotNil(t, s)

	return s.createRouter()
}

func TestNew(t *testing.T) {
	t.Parallel()

	type want struct {
		err  bool
		host string
	}
	tests := []struct {
		name string
		host string
		want want
	}{
		{
			name: "адрес по умолчанию",
			host: "",
			want: want{host: "localhost:8080"},
		},
		{
			name: "заданный адрес",
			host: "127.0.0.1:9090",
			want: want{host: "127.0.0.1:9090"},
		},
		{
			name: "минимальный порт",
			host: "localhost:1",
			want: want{host: "localhost:1"},
		},
		{
			name: "максимальный порт",
			host: "localhost:65535",
			want: want{host: "localhost:65535"},
		},
		{
			name: "без порта",
			host: "localhost",
			want: want{err: true},
		},
		{
			name: "пустой порт",
			host: "localhost:",
			want: want{err: true},
		},
		{
			name: "порт из букв",
			host: "localhost:abc",
			want: want{err: true},
		},
		{
			name: "порт 0",
			host: "localhost:0",
			want: want{err: true},
		},
		{
			name: "порт 65536",
			host: "localhost:65536",
			want: want{err: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s, err := New(test.host)

			if test.want.err {
				assert.Error(t, err)
				assert.Nil(t, s)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, s)
			assert.Equal(t, test.want.host, s.host)
		})
	}
}

func TestRouterUpdate(t *testing.T) {
	t.Parallel()

	type want struct {
		codes []int // допустимые коды ответа
	}
	tests := []struct {
		name   string
		method string
		path   string
		want   want
	}{
		{
			name:   "gauge: корректное значение",
			method: http.MethodPost,
			path:   "/update/gauge/testGauge/100",
			want:   want{codes: []int{http.StatusOK}},
		},
		{
			name:   "gauge: без имени метрики",
			method: http.MethodPost,
			path:   "/update/gauge/",
			want:   want{codes: []int{http.StatusNotFound}},
		},
		{
			name:   "gauge: некорректное значение",
			method: http.MethodPost,
			path:   "/update/gauge/testGauge/none",
			want:   want{codes: []int{http.StatusBadRequest}},
		},
		{
			name:   "counter: корректное значение",
			method: http.MethodPost,
			path:   "/update/counter/testCounter/100",
			want:   want{codes: []int{http.StatusOK}},
		},
		{
			name:   "counter: без имени метрики",
			method: http.MethodPost,
			path:   "/update/counter/",
			want:   want{codes: []int{http.StatusNotFound}},
		},
		{
			name:   "counter: некорректное значение",
			method: http.MethodPost,
			path:   "/update/counter/testCounter/none",
			want:   want{codes: []int{http.StatusBadRequest}},
		},
		{
			name:   "неизвестный тип метрики",
			method: http.MethodPost,
			path:   "/update/unknown/testCounter/100",
			want:   want{codes: []int{http.StatusBadRequest, http.StatusNotImplemented}},
		},
		{
			name:   "неизвестный путь",
			method: http.MethodPost,
			path:   "/updater/counter/testCounter/100",
			want:   want{codes: []int{http.StatusBadRequest, http.StatusNotFound}},
		},
		{
			name:   "GET вместо POST",
			method: http.MethodGet,
			path:   "/update/gauge/testGauge/100",
			want:   want{codes: []int{http.StatusMethodNotAllowed}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := newTestRouter(t)

			code, body := doRequest(t, r, test.method, test.path)

			assert.Contains(t, test.want.codes, code, "тело ответа: %s", body)
		})
	}
}

func TestRouterValue(t *testing.T) {
	t.Parallel()

	type want struct {
		code int
		body string
	}
	tests := []struct {
		name    string
		updates []string
		method  string
		path    string
		want    want
	}{
		{
			name:    "gauge: последнее значение",
			updates: []string{"/update/gauge/g/1", "/update/gauge/g/123.45"},
			method:  http.MethodGet,
			path:    "/value/gauge/g",
			want:    want{code: http.StatusOK, body: "123.45"},
		},
		{
			name:    "counter: значения суммируются",
			updates: []string{"/update/counter/c/5", "/update/counter/c/10", "/update/counter/c/5"},
			method:  http.MethodGet,
			path:    "/value/counter/c",
			want:    want{code: http.StatusOK, body: "20"},
		},
		{
			name:   "неизвестная метрика",
			method: http.MethodGet,
			path:   "/value/gauge/missing",
			want:   want{code: http.StatusNotFound},
		},
		{
			name:    "метрика другого типа",
			updates: []string{"/update/gauge/m/1"},
			method:  http.MethodGet,
			path:    "/value/counter/m",
			want:    want{code: http.StatusNotFound},
		},
		{
			name:   "неизвестный тип",
			method: http.MethodGet,
			path:   "/value/unknown/m",
			want:   want{code: http.StatusNotFound},
		},
		{
			name:   "POST вместо GET",
			method: http.MethodPost,
			path:   "/value/gauge/g",
			want:   want{code: http.StatusMethodNotAllowed},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := newTestRouter(t)

			for _, u := range test.updates {
				code, body := doRequest(t, r, http.MethodPost, u)
				require.Equal(t, http.StatusOK, code, "не удалось выполнить %s: %s", u, body)
			}

			code, body := doRequest(t, r, test.method, test.path)

			assert.Equal(t, test.want.code, code)
			if test.want.body != "" {
				assert.Equal(t, test.want.body, body)
			}
		})
	}
}

func TestRouterAllMetrics(t *testing.T) {
	t.Parallel()

	type want struct {
		code     int
		contains []string
	}
	tests := []struct {
		name    string
		updates []string
		want    want
	}{
		{
			name: "пустое хранилище",
			want: want{
				code:     http.StatusOK,
				contains: []string{"Метрики отсутствуют"},
			},
		},
		{
			name:    "список содержит метрики",
			updates: []string{"/update/gauge/Alloc/123.45", "/update/counter/PollCount/3"},
			want: want{
				code:     http.StatusOK,
				contains: []string{`"Alloc" - 123.45`, `"PollCount" - 3`},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := newTestRouter(t)

			for _, u := range test.updates {
				code, body := doRequest(t, r, http.MethodPost, u)
				require.Equal(t, http.StatusOK, code, "не удалось выполнить %s: %s", u, body)
			}

			code, body := doRequest(t, r, http.MethodGet, "/")

			assert.Equal(t, test.want.code, code)
			for _, s := range test.want.contains {
				assert.Contains(t, body, s)
			}
		})
	}
}

func TestRouterIsolated(t *testing.T) {
	t.Parallel()

	s, err := New("")
	require.NoError(t, err)

	r1, r2 := s.createRouter(), s.createRouter()

	code, _ := doRequest(t, r1, http.MethodPost, "/update/counter/c/1")
	require.Equal(t, http.StatusOK, code)

	code, _ = doRequest(t, r2, http.MethodGet, "/value/counter/c")
	assert.Equal(t, http.StatusNotFound, code)
}
