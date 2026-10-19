package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const envRunApp = "TEST_RUNAPP_ARGS"

func init() {
	if v := os.Getenv(envRunApp); v != "" {
		args := strings.Fields(v)
		poll, _ := strconv.Atoi(args[1])
		report, _ := strconv.Atoi(args[2])
		runApp(args[0], poll, report)
		os.Exit(0)
	}
}

func TestRunApp_ExitsOnInvalidConfig(t *testing.T) {
	t.Parallel()

	type want struct {
		nonZero bool
	}
	tests := []struct {
		name           string
		serverAddr     string
		pollInterval   int
		reportInterval int
		want           want
	}{
		{"отрицательный pollInterval", "localhost:8080", -1, 10, want{true}},
		{"отрицательный reportInterval", "localhost:8080", 2, -1, want{true}},
		{"некорректный адрес", "localhost:abc", 2, 10, want{true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cmd := exec.Command(os.Args[0], "-test.run=^$")
			cmd.Env = append(os.Environ(), envRunApp+"="+strings.Join([]string{
				test.serverAddr,
				strconv.Itoa(test.pollInterval),
				strconv.Itoa(test.reportInterval),
			}, " "))

			out, err := cmd.CombinedOutput()

			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr)
			assert.Equal(t, test.want.nonZero, exitErr.ExitCode() != 0)
			assert.NotEmpty(t, out, "ожидалось сообщение об ошибке")
		})
	}
}

func TestRunApp_ValidConfig(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
	}))
	defer srv.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		runApp(strings.TrimPrefix(srv.URL, "http://"), 1, 1)
	}()

	select {
	case <-done:
		t.Fatal("runApp должен работать до сигнала, но завершился")
	case <-time.After(100 * time.Millisecond):
	}

	require.Eventually(t, func() bool {
		return requests.Load() > 0
	}, 5*time.Second, 50*time.Millisecond, "агент не отправил метрики")

	require.NoError(t, syscall.Kill(os.Getpid(), syscall.SIGTERM))

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("runApp не завершился после SIGTERM")
	}
}
