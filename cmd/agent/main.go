package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/stolyarov-aa/yandex-metrics/internal/agent"
	"github.com/stolyarov-aa/yandex-metrics/internal/service"
)

func runApp(serverAddr string, pollInterval int, reportInterval int) {
	a, err := agent.New(serverAddr, pollInterval, reportInterval)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a.Run(ctx)
}

func main() {
	host := ""
	pollInterval := 0
	reportInterval := 0

	err := service.ParseFlags(&host, &pollInterval, &reportInterval, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	runApp(host, pollInterval, reportInterval)
}
