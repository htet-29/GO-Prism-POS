package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
)

const port = 8089

func TestServerGracefulShutdown(t *testing.T) {
	app := &application{
		config: config{port: port, env: "testing"},
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- app.server()
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", app.config.port))
	if err != nil {
		t.Fatalf("Server failed to start or respond: %v", err)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatalf("Can not close response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find current process: %v", err)
	}

	err = process.Signal(syscall.SIGINT)
	if err != nil {
		t.Fatalf("Failed to send SIGINT: %v", err)
	}

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Expected nil error on clean shutdown, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Test timed out: Server took too long to shut down")
	}
}
