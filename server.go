package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

// Server represents an HTTP server wrapper.
// It contains an underlying http.Server instance and provides methods to start and stop the server.
// It also provides a Run function to start an HTTP server with graceful shutdown.
// The server is stopped gracefully when the context is cancelled or a shutdown signal is received.
type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

// New creates a new instance of the new HTTP server
// with the specified address, handler, and optional server options.
// The server options can be used to customize the server's behavior.
// The addr parameter specifies the address to listen on, e.g., ":8080" for all interfaces on port 8080.
// The handler parameter is an http.Handler that defines the behavior of the server.
// The opt parameter is a variadic list of server options.
// The server options are applied in order, so the last option overrides the previous ones.
// The server options are applied before the server is started.
func New(addr string, handler http.Handler, opt ...serverOption) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr:           addr,
			Handler:        handler,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    15 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1 MB
		},
		shutdownTimeout: 5 * time.Second,
	}

	// Apply options
	for _, o := range opt {
		o(s)
	}

	return s
}

// Start starts the server and listens for incoming requests.
// It uses the provided context to handle graceful shutdown.
// The context is also used to handle shutdown signals from the OS.
// It returns an error if the server fails to start or encounters an error during shutdown.
func (s *Server) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	// Start the server in a new goroutine within the errgroup
	g.Go(func() error {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server ListenAndServe: %w", err)
		}
		return nil
	})

	// Handle shutdown signal in a new goroutine within the errgroup
	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-signalChan():
			return s.Stop(s.shutdownTimeout)
		}
	})

	// Wait for all goroutines in the errgroup to finish
	return g.Wait()
}

// Stop stops the server gracefully with the given timeout.
// It uses the provided timeout to gracefully shutdown the underlying HTTP server.
// If the timeout is reached before the server is fully stopped, an error is returned.
func (s *Server) Stop(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// Close stops the server immediately without waiting for active connections to finish.
// It returns an error if the server fails to stop.
func (s *Server) Close() error {
	return s.httpServer.Close()
}

// signalChan sets up a channel to listen for OS signals for shutdown
func signalChan() <-chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	return stop
}

// Run starts an HTTP server on the specified address and runs the provided handler.
// It returns an error if the server fails to start or encounters an error during execution.
// The context.Context parameter allows for graceful shutdown of the server.
// The addr parameter specifies the address to listen on, e.g., ":8080" for all interfaces on port 8080.
// The handler parameter is an http.Handler that defines the behavior of the server.
func Run(ctx context.Context, addr string, handler http.Handler) error {
	server := New(addr, handler)
	return server.Start(ctx)
}
