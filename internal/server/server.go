package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/stolyarov-aa/yandex-metrics/internal/handler"
	"github.com/stolyarov-aa/yandex-metrics/internal/storage"
)

type Server struct {
	port string
}

func New() *Server {
	return &Server{port: "8080"}
}

func (s *Server) Run() {
	if err := validatePort(s.port); err != nil {
		log.Fatalf("%s", err)
	}

	hub := handler.NewHandlerHub(storage.NewMemStorage())

	r := chi.NewRouter()

	r.Get("/", hub.GetAllMetricsHandler)
	r.Get("/value/{type}/{name}", hub.GetMetricHandler)
	r.Post("/update/{type}/{name}/{value}", hub.UpdateMetricHandler)

	fmt.Printf("Server starting on : %s\n", s.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", s.port), r))
}

func validatePort(port string) error {
	if port == "" {
		return fmt.Errorf("порт не указан")
	}
	num, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("некорректный номер порта %q: не является числом", port)
	}
	if num < 1 || num > 65535 {
		return fmt.Errorf("некорректный номер порта %d: должен быть в диапазоне от 1 до 65535", num)
	}
	return nil
}
