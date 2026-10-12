package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/stolyarov-aa/yandex-metrics/internal/handler"
)

type Server struct {
	port string
}

func New(port string) *Server {
	if err := validatePort(port); err != nil {
		log.Fatalf("%s", err)
	}
	return &Server{port: port}
}

func (s *Server) Run() {
	if err := validatePort(s.port); err != nil {
		log.Fatalf("%s", err)
	}

	r := mux.NewRouter()
	r.StrictSlash(false)

	r.HandleFunc("/update/{type}/{name}/{value}", handler.UpdateMetricHandler).Methods("POST")

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
