package service

import (
	"os"
	"strings"
	"testing"
)

func setEnv(t *testing.T, env map[string]string) func() {
	t.Helper()

	prev := make(map[string]string)
	for k := range env {
		prev[k] = os.Getenv(k)
	}

	for k, v := range env {
		os.Setenv(k, v)
	}

	return func() {
		for k, v := range prev {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}
}

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
			name:    "корректные аргументы",
			host:    &[]string{"localhost:8080"}[0],
			poll:    &[]int{5}[0],
			report:  &[]int{15}[0],
			args:    []string{"-a", "localhost:8080", "-p", "5", "-r", "15"},
			wantErr: false,
		},
		{
			name:    "пустой хост",
			host:    &[]string{""}[0],
			args:    []string{"-a", ""},
			wantErr: true,
			errMsg:  "хост не может быть пустым",
		},
		{
			name:    "неверный pollInterval",
			host:    &[]string{"localhost:8080"}[0],
			poll:    &[]int{0}[0],
			args:    []string{"-p", "0"},
			wantErr: true,
			errMsg:  "pollInterval должен быть > 0",
		},
		{
			name:    "лишние аргументы",
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

func TestParseFlagsWithEnv(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		host        *string
		poll        *int
		report      *int
		args        []string
		wantErr     bool
		checkHost   string
		checkPoll   int
		checkReport int
	}{
		{
			name:      "ADDRESS из окружения",
			env:       map[string]string{"ADDRESS": "example.com:9090"},
			host:      &[]string{"localhost:8080"}[0],
			args:      []string{},
			checkHost: "example.com:9090",
		},
		{
			name:      "POLL_INTERVAL из окружения",
			env:       map[string]string{"POLL_INTERVAL": "7"},
			host:      &[]string{"localhost:8080"}[0],
			poll:      &[]int{2}[0],
			args:      []string{},
			checkPoll: 7,
		},
		{
			name:        "REPORT_INTERVAL из окружения",
			env:         map[string]string{"REPORT_INTERVAL": "20"},
			host:        &[]string{"localhost:8080"}[0],
			report:      &[]int{10}[0],
			args:        []string{},
			checkReport: 20,
		},
		{
			name:    "неверный POLL_INTERVAL",
			env:     map[string]string{"POLL_INTERVAL": "abc"},
			host:    &[]string{"localhost:8080"}[0],
			poll:    &[]int{2}[0],
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "отрицательный POLL_INTERVAL",
			env:     map[string]string{"POLL_INTERVAL": "-5"},
			host:    &[]string{"localhost:8080"}[0],
			poll:    &[]int{2}[0],
			args:    []string{},
			wantErr: true,
		},
		{
			name:      "окружение применяется без аргументов",
			env:       map[string]string{"ADDRESS": "env.com:8080", "POLL_INTERVAL": "3"},
			host:      &[]string{"localhost:8080"}[0],
			poll:      &[]int{2}[0],
			args:      []string{},
			checkHost: "env.com:8080",
			checkPoll: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer setEnv(t, tt.env)()

			err := ParseFlags(tt.host, tt.poll, tt.report, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.checkHost != "" && *tt.host != tt.checkHost {
					t.Errorf("host = %q, want %q", *tt.host, tt.checkHost)
				}
				if tt.checkPoll != 0 && *tt.poll != tt.checkPoll {
					t.Errorf("poll = %d, want %d", *tt.poll, tt.checkPoll)
				}
				if tt.checkReport != 0 && *tt.report != tt.checkReport {
					t.Errorf("report = %d, want %d", *tt.report, tt.checkReport)
				}
			}
		})
	}
}

func TestParseEnvDuration(t *testing.T) {
	tests := []struct {
		name    string
		envName string
		envVal  string
		wantB   bool
		wantVal int
		wantErr bool
	}{
		{
			name:    "пустая переменная",
			envName: "TEST_ENV",
			envVal:  "",
			wantB:   false,
			wantVal: 0,
			wantErr: false,
		},
		{
			name:    "корректное значение",
			envName: "TEST_ENV",
			envVal:  "10",
			wantB:   true,
			wantVal: 10,
			wantErr: false,
		},
		{
			name:    "некорректное значение",
			envName: "TEST_ENV",
			envVal:  "abc",
			wantB:   true,
			wantErr: true,
		},
		{
			name:    "нулевое значение",
			envName: "TEST_ENV",
			envVal:  "0",
			wantB:   true,
			wantErr: true,
		},
		{
			name:    "отрицательное значение",
			envName: "TEST_ENV",
			envVal:  "-5",
			wantB:   true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envName, tt.envVal)
			defer os.Unsetenv(tt.envName)

			b, val, err := parseEnvDuration(tt.envName)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEnvDuration() error = %v, wantErr %v", err, tt.wantErr)
			}
			if b != tt.wantB {
				t.Errorf("parseEnvDuration() b = %v, want %v", b, tt.wantB)
			}
			if !tt.wantErr && val != tt.wantVal {
				t.Errorf("parseEnvDuration() val = %d, want %d", val, tt.wantVal)
			}
		})
	}
}