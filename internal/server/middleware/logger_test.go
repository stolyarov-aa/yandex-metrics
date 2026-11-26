package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	middleware := MiddlewareLogger(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", rec.Code)
	}

	if rec.Body.String() != "test" {
		t.Errorf("ожидалось тело 'test', получено '%s'", rec.Body.String())
	}
}

func TestMiddlewareLoggerWithDifferentStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	middleware := MiddlewareLogger(handler)

	req := httptest.NewRequest("POST", "/not-found", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("ожидался статус 404, получен %d", rec.Code)
	}
}

func TestResponseWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec}

	rw.WriteHeader(http.StatusCreated)
	n, err := rw.Write([]byte("hello"))

	if err != nil {
		t.Errorf("неожиданная ошибка: %v", err)
	}

	if n != 5 {
		t.Errorf("ожидалось 5 байт, получено %d", n)
	}

	if rw.statusCode != http.StatusCreated {
		t.Errorf("ожидался статус 201, получен %d", rw.statusCode)
	}

	if rw.size != 5 {
		t.Errorf("ожидался размер 5, получен %d", rw.size)
	}
}
