package httpserver

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

type serverOption func(*Server)

// WithPreconfiguredServer allows you to provide a preconfigured http.Server instance.
// This is useful if you want to configure TLS or other options that are not exposed by this package.
func WithPreconfiguredServer(s *http.Server) serverOption {
	return func(srv *Server) {
		srv.httpServer = s
	}
}

// WithReadTimeout sets the maximum duration for reading the entire request, including the body.
// This also includes the time spent reading the request header.
// If the server does not receive a new request within this duration it will close the connection.
// A duration of 0 means no timeout.
func WithReadTimeout(d time.Duration) serverOption {
	return func(srv *Server) {
		srv.httpServer.ReadTimeout = d
	}
}

// WithReadHeaderTimeout sets the amount of time allowed to read request headers.
// A duration of 0 means no timeout.
func WithReadHeaderTimeout(d time.Duration) serverOption {
	return func(srv *Server) {
		srv.httpServer.ReadHeaderTimeout = d
	}
}

// WithWriteTimeout sets the maximum duration before timing out writes of the response.
// A duration of 0 means no timeout.
func WithWriteTimeout(d time.Duration) serverOption {
	return func(srv *Server) {
		srv.httpServer.WriteTimeout = d
	}
}

// WithIdleTimeout sets the maximum amount of time to wait for the next request when keep-alives are enabled.
// If IdleTimeout is zero, the value of ReadTimeout is used.
// If both are zero, there is no timeout.
func WithIdleTimeout(d time.Duration) serverOption {
	return func(srv *Server) {
		srv.httpServer.IdleTimeout = d
	}
}

// WithMaxHeaderBytes sets the maximum size of request headers.
// This prevents attacks where an attacker sends a large header to consume server resources.
// If zero, DefaultMaxHeaderBytes of 1MB is used.
func WithMaxHeaderBytes(n int) serverOption {
	return func(srv *Server) {
		srv.httpServer.MaxHeaderBytes = n
	}
}

// WithTLSConfig sets the TLS configuration to use when starting TLS.
// If nil, the default configuration is used.
// If non-nil, HTTP/2 support may not be enabled by default.
func WithTLSConfig(t *tls.Config) serverOption {
	return func(srv *Server) {
		srv.httpServer.TLSConfig = t
	}
}

// WithTLSNextProto sets a function to be called after a TLS handshake has been completed.
// This is useful for protocols which require interaction immediately after the handshake.
// If non-nil, HTTP/2 support may not be enabled by default.
func WithTLSNextProto(f func(*http.Server, *tls.Conn, http.Handler)) serverOption {
	return func(srv *Server) {
		srv.httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
		srv.httpServer.TLSNextProto[""] = f
	}
}

// WithErrorLog sets the error logger for the server.
// If nil, the log package's standard logger is used.
func WithErrorLog(l *log.Logger) serverOption {
	return func(srv *Server) {
		srv.httpServer.ErrorLog = l
	}
}

// WithGracefulShutdown sets the graceful shutdown timeout.
// If zero, the default timeout of 5 seconds is used.
func WithGracefulShutdown(d time.Duration) serverOption {
	return func(srv *Server) {
		srv.shutdownTimeout = d
	}
}

// WithLogger sets the logger for the server.
// If nil, the log package's standard logger is used.
// If you want to use a structured logger, consider using the slog package.
func WithLogger(l Logger) serverOption {
	return func(srv *Server) {
		srv.log = l
	}
}
