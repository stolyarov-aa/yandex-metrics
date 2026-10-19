package service

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
		{"некорректный порт (+123)", "+123", true},
		{"некорректный порт (-123)", "-123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort(%q) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}

func TestValidateHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"пустой хост", "", true},
		{"хост без двоеточия", "localhost", true},
		{"хост с двумя двоеточиями", "localhost:8080:extra", true},
		{"хост с пустым портом", "localhost:", true},
		{"хост с портом 0", "localhost:0", true},
		{"хост с портом > 65535", "localhost:65536", true},
		{"хост с портом буквы", "localhost:abc", true},
		{"хост с портом +123", "localhost:+123", true},
		{"хост с портом -123", "localhost:-123", true},
		{"валидный хост (localhost)", "localhost:8080", false},
		{"валидный хост (0.0.0.0)", "0.0.0.0:8080", false},
		{"валидный хост (IPv6)", "[::1]:8080", false},
		{"валидный хост (min port)", "localhost:1", false},
		{"валидный хост (max port)", "localhost:65535", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHost(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
			}
		})
	}
}
