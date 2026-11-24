package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/stolyarov-aa/yandex-metrics/internal/handler"
	"github.com/stolyarov-aa/yandex-metrics/internal/service"
	"github.com/stolyarov-aa/yandex-metrics/internal/storage"
)

type Server struct {
	host string
}

func New(host string) (*Server, error) {
	if host == "" {
		host = "localhost:8080"
	}

	if err := service.ValidateHost(host); err != nil {
		return nil, err
	}

	return &Server{host: host}, nil
}

func (s *Server) createRouter() http.Handler {
	hub := handler.NewHandlerHub(storage.NewMemStorage())

	r := chi.NewRouter()

	r.Get("/", hub.GetAllMetricsHandler)
	r.Get("/value/{type}/{name}", hub.GetMetricHandler)
	r.Post("/update/{type}/{name}/{value}", hub.UpdateMetricHandler)

	return r
}

func (s *Server) Run() {
	fmt.Printf("Server starting on : %s\n", s.host)
	log.Fatal(http.ListenAndServe(s.host, s.createRouter()))
}
