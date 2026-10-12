package main

import (
	"github.com/stolyarov-aa/yandex-metrics/internal/server"
)

func main() {
	srv := server.New("8080")
	srv.Run()
}
