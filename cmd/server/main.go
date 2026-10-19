package main

import (
	"log"
	"os"

	"github.com/stolyarov-aa/yandex-metrics/internal/server"
	"github.com/stolyarov-aa/yandex-metrics/internal/service"
)

func main() {
	host := ""

	err := service.ParseFlags(&host, nil, nil, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	srv, err := server.New(host)
	if err != nil {
		log.Fatal(err)
	}

	srv.Run()
}
