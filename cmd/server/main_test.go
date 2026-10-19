package main

import (
	"bytes"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	envRun  = "TEST_SERVER_RUN_MAIN"
	envArgs = "TEST_SERVER_ARGS"
)

func init() {
	if os.Getenv(envRun) == "1" {
		os.Args = append([]string{"server"}, strings.Fields(os.Getenv(envArgs))...)
		main()
		os.Exit(0)
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func startMain(t *testing.T, args ...string) (*syncBuffer, chan error) {
	t.Helper()

	stderr := &syncBuffer{}
	done := make(chan error, 1)

	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), envRun+"=1", envArgs+"="+strings.Join(args, " "))
	cmd.Stderr = stderr
	require.NoError(t, cmd.Start())

	go func() { done <- cmd.Wait() }()
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
	})
	return stderr, done
}

func freeAddr(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	return l.Addr().String()
}

func TestMainExitsOnError(t *testing.T) {
	t.Parallel()

	type want struct {
		nonZero bool
	}
	tests := []struct {
		name string
		args []string
		want want
	}{
		{name: "неизвестный флаг", args: []string{"-x", "1"}, want: want{nonZero: true}},
		{name: "лишний аргумент", args: []string{"extra"}, want: want{nonZero: true}},
		{name: "некорректный адрес", args: []string{"-a", "localhost:abc"}, want: want{nonZero: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			stderr, done := startMain(t, test.args...)

			select {
			case err := <-done:
				var exitErr *exec.ExitError
				require.ErrorAs(t, err, &exitErr)
				assert.Equal(t, test.want.nonZero, exitErr.ExitCode() != 0)
			case <-time.After(5 * time.Second):
				require.FailNow(t, "процесс не завершился")
			}
			assert.NotEmpty(t, stderr.String(), "ожидалось сообщение об ошибке")
		})
	}
}

func TestMainStartsServer(t *testing.T) {
	t.Parallel()

	addr := freeAddr(t)
	stderr, _ := startMain(t, "-a", addr)

	ok := assert.Eventually(t, func() bool {
		resp, err := http.Get("http://" + addr + "/")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 50*time.Millisecond, "сервер не запустился")
	if !ok {
		t.Fatalf("лог сервера:\n%s", stderr.String())
	}

	resp, err := http.Post("http://"+addr+"/update/counter/c/1", "text/plain", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}