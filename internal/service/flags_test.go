package service

import (
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		host    *string
		poll    *int
		report  *int
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid args",
			host:    &[]string{"localhost:8080"}[0],
			poll:    &[]int{5}[0],
			report:  &[]int{15}[0],
			args:    []string{"-a", "localhost:8080", "-p", "5", "-r", "15"},
			wantErr: false,
		},
		{
			name:    "empty host",
			host:    &[]string{""}[0],
			args:    []string{"-a", ""},
			wantErr: true,
			errMsg:  "хост не может быть пустым",
		},
		{
			name:    "invalid pollInterval",
			host:    &[]string{"localhost:8080"}[0],
			poll:    &[]int{0}[0],
			args:    []string{"-p", "0"},
			wantErr: true,
			errMsg:  "pollInterval должен быть > 0",
		},
		{
			name:    "extra args",
			host:    &[]string{"localhost:8080"}[0],
			args:    []string{"-a", "localhost:8080", "extra"},
			wantErr: true,
			errMsg:  "неизвестные аргументы",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseFlags(tt.host, tt.poll, tt.report, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errMsg != "" && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ParseFlags() error = %q, want to contain %q", err, tt.errMsg)
			}
		})
	}
}
