package main

import (
	"testing"
	"time"
)

func TestRunApp_PanicOnInvalidConfig(t *testing.T) {
	tests := []struct {
		name           string
		serverAddr     string
		pollInterval   time.Duration
		reportInterval time.Duration
	}{
		{"пустой serverAddr", "", 2 * time.Second, 10 * time.Second},
		{"ноль pollInterval", "http://localhost:8080", 0, 10 * time.Second},
		{"ноль reportInterval", "http://localhost:8080", 2 * time.Second, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Нет ожидаемой паники")
				}
			}()

			runApp(tt.serverAddr, tt.pollInterval, tt.reportInterval)
		})
	}
}

func TestRunApp_ValidConfig(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("неожиданная паника: %v", r)
			}
		}()

		runApp("http://localhost:8080", 1*time.Millisecond, 2*time.Millisecond)
	}()

	select {
	case <-done:
		t.Errorf("runApp должен работать бесконечно, но завершился")
	case <-time.After(100 * time.Millisecond):
	}
}
