package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stolyarov-aa/yandex-metrics/internal/agent"
)

func runApp(serverAddr string, pollInterval, reportInterval time.Duration) {
	a := agent.New(serverAddr, pollInterval, reportInterval)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a.Run(ctx)
}

func main() {
	runApp(
		"http://localhost:8080",
		2*time.Second,
		10*time.Second,
	)
}
