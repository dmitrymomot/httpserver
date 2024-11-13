package httpserver_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/dmitrymomot/httpserver"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	listenAddr := "localhost:9999"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})
	server, err := httpserver.New(listenAddr, handler)
	require.NoError(t, err, "Unexpected error creating server")

	// Create a context with cancel for server control
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to catch server errors
	serverErr := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		serverErr <- server.Start(ctx)
	}()

	// Wait for the server to start
	time.Sleep(500 * time.Millisecond)

	// Test server response
	resp, err := http.Get(fmt.Sprintf("http://%s", listenAddr))
	require.NoError(t, err, "Unexpected error in GET request")
	require.Equal(t, http.StatusOK, resp.StatusCode, "Unexpected status code")
	resp.Body.Close()

	// Initiate graceful shutdown
	cancel()

	// Wait for server to shut down with timeout
	shutdownTimeout := time.After(5 * time.Second)
	select {
	case err := <-serverErr:
		require.True(t, err == nil || errors.Is(err, context.Canceled),
			"Expected nil or context.Canceled error, got: %v", err)
	case <-shutdownTimeout:
		t.Fatal("Server shutdown timed out")
	}

	// Verify server is no longer accepting connections
	_, err = http.Get(fmt.Sprintf("http://%s", listenAddr))
	require.Error(t, err, "Expected error after server shutdown")
}
