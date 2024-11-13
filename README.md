# httpserver

[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/dmitrymomot/httpserver)](https://github.com/dmitrymomot/httpserver)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmitrymomot/httpserver.svg)](https://pkg.go.dev/github.com/dmitrymomot/httpserver)
[![License](https://img.shields.io/github/license/dmitrymomot/httpserver)](https://github.com/dmitrymomot/httpserver/blob/main/LICENSE)

[![Tests](https://github.com/dmitrymomot/httpserver/actions/workflows/tests.yml/badge.svg)](https://github.com/dmitrymomot/httpserver/actions/workflows/tests.yml)
[![CodeQL Analysis](https://github.com/dmitrymomot/httpserver/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/dmitrymomot/httpserver/actions/workflows/codeql-analysis.yml)
[![GolangCI Lint](https://github.com/dmitrymomot/httpserver/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/dmitrymomot/httpserver/actions/workflows/golangci-lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmitrymomot/httpserver)](https://goreportcard.com/report/github.com/dmitrymomot/httpserver)

The `httpserver` package provides a robust and feature-rich HTTP server implementation in Go, offering graceful shutdown, static file serving, and extensive configuration options. It's designed to be both simple to use and powerful enough for production environments.

## Features

-   Easy server setup and configuration
-   Graceful shutdown with configurable timeout
-   Context-based cancellation
-   Static file serving with caching support
-   Embedded filesystem support
-   Comprehensive server options
-   Structured logging support
-   TLS configuration
-   Concurrent execution with `errgroup`
-   Production-ready defaults

## Installation

```bash
go get github.com/dmitrymomot/httpserver
```

## Basic Usage

### Simple HTTP Server

```go
package main

import (
    "context"
    "net/http"
    "github.com/dmitrymomot/httpserver"
)

func main() {
    // Create a new router
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Start the server with default configuration
    if err := httpserver.Run(context.Background(), ":8080", mux); err != nil {
        panic(err)
    }
}
```

### Advanced Server Configuration

```go
package main

import (
    "context"
    "net/http"
    "time"
    "github.com/dmitrymomot/httpserver"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Create a new server with custom options
    server, err := httpserver.New(":8080", mux,
        httpserver.WithReadTimeout(5*time.Second),
        httpserver.WithWriteTimeout(10*time.Second),
        httpserver.WithIdleTimeout(15*time.Second),
        httpserver.WithGracefulShutdown(30*time.Second),
        httpserver.WithMaxHeaderBytes(1<<20), // 1MB
    )
    if err != nil {
        panic(err)
    }

    // Start the server with context
    ctx := context.Background()
    if err := server.Start(ctx); err != nil {
        panic(err)
    }
}
```

### Serving Static Files

```go
package main

import (
    "context"
    "net/http"
    "time"
    "github.com/dmitrymomot/httpserver"
)

func main() {
    mux := http.NewServeMux()

    // Serve files from physical directory
    mux.HandleFunc("/static/", httpserver.StaticHandler(
        "/static",
        http.Dir("./static"),
        10*time.Minute, // Cache TTL
    ))

    // Serve files from embedded filesystem
    //go:embed static/*
    var embedFS embed.FS
    mux.HandleFunc("/assets/", httpserver.EmbeddedStaticHandler(
        embedFS,
        24*time.Hour, // Cache TTL
    ))

    if err := httpserver.Run(context.Background(), ":8080", mux); err != nil {
        panic(err)
    }
}
```

### Graceful Shutdown

```go
package main

import (
    "context"
    "net/http"
    "time"
    "github.com/dmitrymomot/httpserver"
)

func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    mux := http.NewServeMux()
    server, _ := httpserver.New(":8080", mux)

    // Server will shut down gracefully when context is canceled
    if err := server.Start(ctx); err != nil {
        panic(err)
    }
}
```

## Server Options

The package provides numerous options to configure the server:

-   `WithPreconfiguredServer` - Use a pre-configured http.Server
-   `WithReadTimeout` - Set maximum duration for reading requests
-   `WithWriteTimeout` - Set maximum duration for writing responses
-   `WithIdleTimeout` - Set maximum time to wait for the next request
-   `WithMaxHeaderBytes` - Set maximum size of request headers
-   `WithTLSConfig` - Configure TLS settings
-   `WithGracefulShutdown` - Set graceful shutdown timeout
-   `WithLogger` - Set custom logger

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the [Apache 2.0](LICENSE) - see the LICENSE file for details.
