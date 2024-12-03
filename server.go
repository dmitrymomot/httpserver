package httpserver

import (
	"context"
	"errors"
	"log/slog"
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
	log             Logger
}

// Logger is an interface that defines the logging methods used by the server.
// It is used to log messages during server startup and shutdown.
// The server uses the Logger interface to log messages with context information.
type Logger interface {
	InfoContext(ctx context.Context, msg string, keyvals ...interface{})
	ErrorContext(ctx context.Context, msg string, keyvals ...interface{})
}

// New creates a new instance of the new HTTP server
// with the specified address, handler, and optional server options.
// The server options can be used to customize the server's behavior.
// The addr parameter specifies the address to listen on, e.g., ":8080" for all interfaces on port 8080.
// The handler parameter is an http.Handler that defines the behavior of the server.
// The opt parameter is a variadic list of server options.
// The server options are applied in order, so the last option overrides the previous ones.
// The server options are applied before the server is started.
func New(addr string, handler http.Handler, opt ...serverOption) (*Server, error) {
	if addr == "" {
		return nil, ErrEmptyAddress
	}
	if handler == nil {
		return nil, ErrNilHandler
	}

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
		log:             slog.Default().With(slog.String("component", "httpserver")),
	}

	// Apply options
	for _, o := range opt {
		o(s)
	}

	return s, nil
}

// Start starts the server and listens for incoming requests.
// It uses the provided context to handle graceful shutdown.
// The context is also used to handle shutdown signals from the OS.
// It returns an error if the server fails to start or encounters an error during shutdown.
func (s *Server) Start(ctx context.Context) error {
	s.log.InfoContext(ctx, "starting HTTP server",
		"addr", s.httpServer.Addr,
		"read_timeout", s.httpServer.ReadTimeout,
		"write_timeout", s.httpServer.WriteTimeout,
		"idle_timeout", s.httpServer.IdleTimeout,
	)

	// Create a new context for shutdown
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	// creates ctx which will be canceled on first failed goroutine
	g, ctx := errgroup.WithContext(ctx)

	// Start the server in a new goroutine within the errgroup
	g.Go(func() error {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return errors.Join(ErrServerStart, err)
		}
		return nil
	})

	// Handle shutdown signals
	g.Go(func() error {
		select {
		case <-ctx.Done():
			s.log.InfoContext(ctx, "context cancelled, initiating shutdown")
			return s.Stop(shutdownCtx, s.shutdownTimeout)
		case sig := <-signalChan():
			s.log.InfoContext(ctx, "received shutdown signal", "signal", sig.String())
			return s.Stop(shutdownCtx, s.shutdownTimeout)
		}
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		s.log.ErrorContext(ctx, "server stopped with error", "error", err)
		return err
	}

	s.log.InfoContext(ctx, "server stopped gracefully")
	return nil
}

// Stop stops the server gracefully with the given timeout.
// It uses the provided timeout to gracefully shutdown the underlying HTTP server.
// If the timeout is reached before the server is fully stopped, an error is returned.
func (s *Server) Stop(ctx context.Context, timeout time.Duration) error {
	s.log.InfoContext(ctx, "stopping HTTP server", "timeout", timeout)

	// Create a new context for shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create an error group for coordinated shutdown
	g := new(errgroup.Group)

	// Shutdown the HTTP server
	g.Go(func() error {
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil &&
			!errors.Is(err, context.Canceled) &&
			!errors.Is(err, http.ErrServerClosed) {
			return errors.Join(ErrServerStop, err)
		}
		return nil
	})

	// Wait for shutdown to complete or timeout
	if err := g.Wait(); err != nil {
		s.log.ErrorContext(ctx, "error during server shutdown", "error", err)
		// Force close if graceful shutdown fails
		_ = s.Close(ctx)
		return err
	}

	s.log.InfoContext(ctx, "HTTP server shutdown complete")
	return nil
}

// Close stops the server immediately without waiting for active connections to finish.
// It returns an error if the server fails to stop.
func (s *Server) Close(ctx context.Context) error {
	s.log.InfoContext(ctx, "force closing HTTP server")

	if err := s.httpServer.Close(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.ErrorContext(ctx, "error during force close", "error", err)
		return errors.Join(ErrServerForceClose, err)
	}
	return nil
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
	server, err := New(addr, handler)
	if err != nil {
		return err
	}
	return server.Start(ctx)
}
