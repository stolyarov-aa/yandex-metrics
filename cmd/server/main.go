package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/stolyarov-aa/yandex-metrics/internal/server"
	"github.com/stolyarov-aa/yandex-metrics/internal/service"
)

func main() {
	host := ""

	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

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
