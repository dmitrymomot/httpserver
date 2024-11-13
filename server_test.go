package httpserver_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/dmitrymomot/httpserver"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	listenAddr := "localhost:9999"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "Hello, World!") })
	server := httpserver.New(listenAddr, handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the server in a separate goroutine
	go func() {
		if err := server.Start(ctx); err != nil {
			require.NoError(t, err, "Unexpected error in server start")
		}
	}()

	// Wait for the server to start
	time.Sleep(500 * time.Millisecond)

	// Perform an HTTP request to the server
	resp, err := http.Get(fmt.Sprintf("http://%s", listenAddr))
	require.NoError(t, err, "Unexpected error in GET request")
	require.Equal(t, http.StatusOK, resp.StatusCode, "Unexpected status code")

	// Shutdown the server
	require.NoError(t, server.Stop(ctx, 1*time.Second), "Unexpected error in server shutdown")

	// Wait for the server to shut down
	// time.Sleep(1 * time.Second)

	// Perform an HTTP request to the server after it has shut down
	_, err = http.Get(fmt.Sprintf("http://%s", listenAddr))
	require.Error(t, err, "Expected error after server shutdown")
}
