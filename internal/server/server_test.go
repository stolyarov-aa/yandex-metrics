package server

import (
	"testing"
)

func TestServerNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		port   string
		wantOK bool
	}{
		{"пустой порт", "", false},
		{"некорректный порт (буквы)", "abc", false},
		{"некорректный порт (0)", "0", false},
		{"некорректный порт (65536)", "65536", false},
		{"валидный порт", "8080", true},
		{"валидный порт (min)", "1", true},
		{"валидный порт (max)", "65535", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil && !tt.wantOK {
					t.Errorf("ожидалась паника, но её не было")
				} else if r != nil && tt.wantOK {
					t.Errorf("неожиданная паника: %v", r)
				}
			}()

			srv := New(tt.port)
			if srv == nil && tt.wantOK {
				t.Errorf("ожидался сервер, получено nil")
			}
		})
	}
}

func TestServerRun(t *testing.T) {
	t.Parallel()

	srv := New("0")
	if srv == nil {
		t.Fatal("не удалось создать сервер")
	}

	// Запускаем сервер в goroutine
	done := make(chan struct{})
	go func() {
		srv.Run()
		close(done)
	}()

	t.Log("Сервер запущен и слушает на случайном порту")
}
