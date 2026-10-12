package server

import (
	"testing"
)

func TestValidatePort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"пустой порт", "", true},
		{"некорректный порт (буквы)", "abc", true},
		{"некорректный порт (0)", "0", true},
		{"некорректный порт (65536)", "65536", true},
		{"валидный порт", "8080", false},
		{"валидный порт (min)", "1", false},
		{"валидный порт (max)", "65535", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePort(%q) error = %v, wantErr = %v", tt.port, err, tt.wantErr)
			}
		})
	}
}

func TestServerNew(t *testing.T) {
	t.Parallel()

	srv := New()
	if srv == nil {
		t.Fatal("ожидался сервер, получено nil")
	}
	if srv.port != "8080" {
		t.Errorf("ожидался порт 8080, получено %s", srv.port)
	}
}
